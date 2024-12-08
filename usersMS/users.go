package usersMS

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
	_ "github.com/go-sql-driver/mysql"
)

var jwtKey = []byte("f0a567902063bf0d5d1e6c0c73a4666e70698882aa9b0ae508c7265efdc52865399728aee1a08bb17ef5de60fc8828ade449fde6f8ca1ed30244fef462c4f39f74c37dc80d1dea2873e7cee198d8a333ea5fe4b494c2f3bf9dc5e535cd8812442a0b6b4d7a93d47fb428e3320ec8448c314c576fabbbadb593489299b877d2a5557d1ef5dca07231b02f01c0a0ba8b043975ed38e81f736761b6e3db3a54847da5f1f29ec31426c0aa10308ea1ba35575b4b936d3cc6903e86afa33a3539ce9b437b6d22a04359061276a8a4cc9054c14a8db10625f07ab9d1ce345d91b717519923887dc4ad05f90564b0bb6dfc2a0e0da3cd4741254b4ccb0f8d267ebf14d6")

var db *sql.DB

type User struct {
	ID           int
	Email        string
	PasswordHash string
	Membership   string
}

func InitUserService() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/users_db")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection is not active: %v", err)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		log.Printf("Error parsing template: %v", err)
		http.Error(w, "Error rendering template", http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func generateJWT(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(jwtKey)
}

// User registration
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		membership := r.FormValue("membership")

		hashedPassword, err := hashPassword(password)
		if err != nil {
			log.Printf("Error hashing password for email %s: %v", email, err)
			http.Error(w, fmt.Sprintf("Error hashing password: %v", err), http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users (email, password, membership) VALUES (?, ?, ?)", email, hashedPassword, membership)
		if err != nil {
			log.Printf("Error creating user with email %s: %v", email, err)
			http.Error(w, fmt.Sprintf("Error creating user: %v", err), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		renderTemplate(w, "templates/register.html", nil)
	}
}

// User login
func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var user User
		err := db.QueryRow("SELECT id, email, password, membership FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Membership)
		if err != nil {
			log.Printf("Login failed for email %s: %v", email, err)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		if !checkPasswordHash(password, user.PasswordHash) {
			log.Printf("Invalid password attempt for email %s", email)
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		token, err := generateJWT(email)
		if err != nil {
			log.Printf("Error generating token for email %s: %v", email, err)
			http.Error(w, fmt.Sprintf("Error generating token: %v", err), http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
	} else {
		renderTemplate(w, "templates/login.html", nil)
	}
}

// Update User Details
func UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		log.Printf("Error retrieving cookie: %v", err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse and validate JWT
	claims := &jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		log.Printf("Error parsing JWT for cookie %s: %v", cookie.Value, err)
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := (*claims)["email"].(string)

	// Fetch the current user from the database
	var user User
	err = db.QueryRow("SELECT id, email, membership FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.Membership)
	if err != nil {
		log.Printf("Error fetching user %s from database: %v", email, err)
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodPost {
		// Get the new details from the form
		newEmail := r.FormValue("email")
		newPassword := r.FormValue("password")
		newMembership := r.FormValue("membership")

		var hashedPassword string
		if newPassword != "" {
			hashedPassword, err = hashPassword(newPassword)
			if err != nil {
				log.Printf("Error hashing new password for user %s: %v", email, err)
				http.Error(w, fmt.Sprintf("Error hashing password: %v", err), http.StatusInternalServerError)
				return
			}
		}

		// Update user details in the database
		if newEmail != "" && newEmail != user.Email {
			_, err = db.Exec("UPDATE users SET email = ? WHERE id = ?", newEmail, user.ID)
			if err != nil {
				log.Printf("Error updating email for user %s: %v", email, err)
				http.Error(w, fmt.Sprintf("Error updating email: %v", err), http.StatusInternalServerError)
				return
			}
			user.Email = newEmail 
		}

		if newPassword != "" {
			_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, user.ID)
			if err != nil {
				log.Printf("Error updating password for user %s: %v", email, err)
				http.Error(w, fmt.Sprintf("Error updating password: %v", err), http.StatusInternalServerError)
				return
			}
		}

		if newMembership != "" && newMembership != user.Membership {
			_, err = db.Exec("UPDATE users SET membership = ? WHERE id = ?", newMembership, user.ID)
			if err != nil {
				log.Printf("Error updating membership for user %s: %v", email, err)
				http.Error(w, fmt.Sprintf("Error updating membership: %v", err), http.StatusInternalServerError)
				return
			}
			user.Membership = newMembership
		}

		http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
	}

	// Render the update form
	renderTemplate(w, "templates/update_user.html", user)
}

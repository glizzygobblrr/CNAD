package main

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

var jwtKey = []byte("f0a567902063bf0d5d1e6c0c73a4666e70698882aa9b0ae508c7265efdc52865399728aee1a08bb17ef5de60fc8828ade449fde6f8ca1ed30244fef462c4f39f74c37dc80d1dea2873e7cee198d8a333ea5fe4b494c2f3bf9dc5e535cd8812442a0b6b4d7a93d47fb428e3320ec8448c314c576fabbbadb593489299b877d2a5557d1ef5dca07231b02f01c0a0ba8b043975ed38e81f736761b6e3db3a54847da5f1f29ec31426c0aa10308ea1ba35575b4b936d3cc6903e86afa33a3539ce9b437b6d22a04359061276a8a4cc9054c14a8db10625f07ab9d1ce345d91b717519923887dc4ad05f90564b0bb6dfc2a0e0da3cd4741254b4ccb0f8d267ebf14d6") // Change this to a secure key
var db *sql.DB

// User struct
type User struct {
	ID           int
	Email        string
	PasswordHash string
	Membership   string
}

// Vehicle struct
type Vehicle struct {
	ID   int
	Name string
}

// Booking struct
type Booking struct {
    ID        int
    Vehicle   string
    StartTime string
    EndTime   string
    Price     float64 
}


// Initialize database
func initDB() {
	var err error
	db, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/carshare")
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("Database connection is not active: %v", err)
	}
}

// Render HTML templates
func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

// Hash password
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

// Check password
func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

// Generate JWT
func generateJWT(email string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"email": email,
		"exp":   time.Now().Add(24 * time.Hour).Unix(),
	})
	return token.SignedString(jwtKey)
}

// User registration
func registerHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")
		membership := r.FormValue("membership")

		hashedPassword, err := hashPassword(password)
		if err != nil {
			http.Error(w, "Error hashing password", http.StatusInternalServerError)
			return
		}

		_, err = db.Exec("INSERT INTO users (email, password, membership) VALUES (?, ?, ?)", email, hashedPassword, membership)
		if err != nil {
			http.Error(w, "Error creating user", http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	} else {
		renderTemplate(w, "templates/register.html", nil)
	}
}

// User login
func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		email := r.FormValue("email")
		password := r.FormValue("password")

		var user User
		err := db.QueryRow("SELECT id, email, password, membership FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.PasswordHash, &user.Membership)
		if err != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		if !checkPasswordHash(password, user.PasswordHash) {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		token, err := generateJWT(email)
		if err != nil {
			http.Error(w, "Error generating token", http.StatusInternalServerError)
			return
		}

		http.SetCookie(w, &http.Cookie{
			Name:     "token",
			Value:    token,
			Expires:  time.Now().Add(24 * time.Hour),
			HttpOnly: true,
		})
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	} else {
		renderTemplate(w, "templates/login.html", nil)
	}
}

// Update User Details
func updateUserHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse and validate JWT
	claims := &jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := (*claims)["email"].(string)

	// Fetch the current user from the database
	var user User
	err = db.QueryRow("SELECT id, email, membership FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.Membership)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Handle form submission
	if r.Method == http.MethodPost {
		// Get the new details from the form
		newEmail := r.FormValue("email")
		newPassword := r.FormValue("password")
		newMembership := r.FormValue("membership")

		var hashedPassword string
		if newPassword != "" {
			hashedPassword, err = hashPassword(newPassword)
			if err != nil {
				http.Error(w, "Error hashing password", http.StatusInternalServerError)
				return
			}
		}

		// Update user details in the database
		if newEmail != "" && newEmail != user.Email {
			_, err = db.Exec("UPDATE users SET email = ? WHERE id = ?", newEmail, user.ID)
			if err != nil {
				http.Error(w, "Error updating email", http.StatusInternalServerError)
				return
			}
			user.Email = newEmail // Update in the local variable
		}

		if newPassword != "" {
			_, err = db.Exec("UPDATE users SET password = ? WHERE id = ?", hashedPassword, user.ID)
			if err != nil {
				http.Error(w, "Error updating password", http.StatusInternalServerError)
				return
			}
		}

		if newMembership != "" && newMembership != user.Membership {
			_, err = db.Exec("UPDATE users SET membership = ? WHERE id = ?", newMembership, user.ID)
			if err != nil {
				http.Error(w, "Error updating membership", http.StatusInternalServerError)
				return
			}
			user.Membership = newMembership // Update in the local variable
		}

		// Redirect to the dashboard or profile page
		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}

	// Render the update form with current user details pre-filled
	renderTemplate(w, "templates/update_user.html", user)
}

// Dashboard
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	tokenStr := cookie.Value
	claims := &jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil || !token.Valid {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := (*claims)["email"].(string)
	var user User
	err = db.QueryRow("SELECT id, email, membership FROM users WHERE email = ?", email).Scan(&user.ID, &user.Email, &user.Membership)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	renderTemplate(w, "templates/dashboard.html", user)
}

// View available vehicles
func viewVehiclesHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse and validate JWT
	claims := &jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Fetch available vehicles
	rows, err := db.Query("SELECT id, name FROM vehicles WHERE is_available = TRUE")
	if err != nil {
		http.Error(w, "Error fetching vehicles", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var vehicles []Vehicle

	for rows.Next() {
		var vehicle Vehicle
		if err := rows.Scan(&vehicle.ID, &vehicle.Name); err != nil {
			http.Error(w, "Error scanning vehicles", http.StatusInternalServerError)
			return
		}
		vehicles = append(vehicles, vehicle)
	}

	renderTemplate(w, "templates/vehicles.html", vehicles)
}

func calculatePrice(membership string, startTime, endTime string) float64 {
    duration := calculateDuration(startTime, endTime) // Calculate rental duration
    basePrice := 10.0 // Default base price per hour

    if membership == "Premium" {
        return basePrice * 0.9 * duration // 10% discount for Premium members
    }
    return basePrice * duration // Basic members pay the full price
}

func calculateDuration(startTime, endTime string) float64 {
    start, _ := time.Parse("2006-01-02 15:04:05", startTime)
    end, _ := time.Parse("2006-01-02 15:04:05", endTime)
    return end.Sub(start).Hours() // Duration in hours
}

// Book a vehicle
func bookVehicleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Parse and validate JWT
		claims := &jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		email := (*claims)["email"].(string)
		var userID int
		var membership string
		err = db.QueryRow("SELECT id, membership FROM users WHERE email = ?", email).Scan(&userID, &membership)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		vehicleID := r.FormValue("vehicle_id")
		startTime := r.FormValue("start_time")
		endTime := r.FormValue("end_time")

		// Calculate price based on membership
		price := calculatePrice(membership, startTime, endTime)

		// Insert booking
		_, err = db.Exec(
			"INSERT INTO bookings (user_id, vehicle_id, start_time, end_time, price) VALUES (?, ?, ?, ?, ?)",
			userID, vehicleID, startTime, endTime, price,
		)
		if err != nil {
			http.Error(w, "Error creating booking", http.StatusInternalServerError)
			return
		}

		// Mark vehicle as unavailable
		_, err = db.Exec("UPDATE vehicles SET is_available = FALSE WHERE id = ?", vehicleID)
		if err != nil {
			http.Error(w, "Error updating vehicle availability", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/dashboard", http.StatusSeeOther)
	}
}


// View user bookings
func viewBookingsHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("token")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Parse and validate JWT
	claims := &jwt.MapClaims{}
	_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	email := (*claims)["email"].(string)
	var userID int
	err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Fetch user bookings
	rows, err := db.Query(
		`SELECT b.id, v.name, b.start_time, b.end_time 
		FROM bookings b 
		JOIN vehicles v ON b.vehicle_id = v.id 
		WHERE b.user_id = ?`, userID,
	)
	if err != nil {
		http.Error(w, "Error fetching bookings", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var bookings []Booking

	for rows.Next() {
		var booking Booking
		if err := rows.Scan(&booking.ID, &booking.Vehicle, &booking.StartTime, &booking.EndTime); err != nil {
			http.Error(w, "Error scanning bookings", http.StatusInternalServerError)
			return
		}
		bookings = append(bookings, booking)
	}

	renderTemplate(w, "templates/bookings.html", bookings)
}

// Cancel a booking
func cancelBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Parse and validate JWT
		claims := &jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		email := (*claims)["email"].(string)
		var userID int
		err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		// Get booking ID from form data (POST)
		bookingID := r.FormValue("booking_id")
		if bookingID == "" {
			http.Error(w, "Booking ID is required", http.StatusBadRequest)
			return
		}

		// Fetch booking details
		var vehicleID int
		var startTime, endTime string
		err = db.QueryRow("SELECT vehicle_id, start_time, end_time FROM bookings WHERE id = ? AND user_id = ?", bookingID, userID).Scan(&vehicleID, &startTime, &endTime)
		if err != nil {
			http.Error(w, "Booking not found or you don't have permission to cancel this booking", http.StatusForbidden)
			return
		}

		// Cancel booking: Delete the booking record
		_, err = db.Exec("DELETE FROM bookings WHERE id = ?", bookingID)
		if err != nil {
			http.Error(w, "Error canceling booking", http.StatusInternalServerError)
			return
		}

		// Make the vehicle available again
		_, err = db.Exec("UPDATE vehicles SET is_available = TRUE WHERE id = ?", vehicleID)
		if err != nil {
			http.Error(w, "Error updating vehicle availability", http.StatusInternalServerError)
			return
		}

		// Redirect to dashboard after canceling
		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
	}
}

// Update a booking
func updateBookingHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Parse and validate JWT
		claims := &jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		email := (*claims)["email"].(string)
		var userID int
		err = db.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		// Get the booking ID and new details from the form data
		bookingID := r.FormValue("booking_id")
		startTime := r.FormValue("start_time")
		endTime := r.FormValue("end_time")

		// Ensure booking ID and times are provided
		if bookingID == "" || startTime == "" || endTime == "" {
			http.Error(w, "Booking ID, start time, and end time are required", http.StatusBadRequest)
			return
		}

		// Check if the booking exists and belongs to the logged-in user
		_, err = db.Exec(
			"UPDATE bookings SET start_time = ?, end_time = ? WHERE id = ? AND user_id = ?",
			startTime, endTime, bookingID, userID,
		)
		if err != nil {
			http.Error(w, "Error updating booking", http.StatusInternalServerError)
			return
		}

		// Redirect to the bookings page after updating
		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
	}
}


func main() {
	initDB()
	defer db.Close()

	http.HandleFunc("/register", registerHandler)
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/update_user", updateUserHandler)
	http.HandleFunc("/dashboard", dashboardHandler)

	http.HandleFunc("/vehicles", viewVehiclesHandler)
	http.HandleFunc("/book", bookVehicleHandler)
	http.HandleFunc("/bookings", viewBookingsHandler)
	http.HandleFunc("/cancel_booking", cancelBookingHandler)
	http.HandleFunc("/update_booking", updateBookingHandler)

	fmt.Println("Server started at http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

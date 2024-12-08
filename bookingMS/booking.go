package bookingMS

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"CNAD/billingMS"
	_ "github.com/go-sql-driver/mysql"
	"github.com/dgrijalva/jwt-go"
)


var jwtKey = []byte("f0a567902063bf0d5d1e6c0c73a4666e70698882aa9b0ae508c7265efdc52865399728aee1a08bb17ef5de60fc8828ade449fde6f8ca1ed30244fef462c4f39f74c37dc80d1dea2873e7cee198d8a333ea5fe4b494c2f3bf9dc5e535cd8812442a0b6b4d7a93d47fb428e3320ec8448c314c576fabbbadb593489299b877d2a5557d1ef5dca07231b02f01c0a0ba8b043975ed38e81f736761b6e3db3a54847da5f1f29ec31426c0aa10308ea1ba35575b4b936d3cc6903e86afa33a3539ce9b437b6d22a04359061276a8a4cc9054c14a8db10625f07ab9d1ce345d91b717519923887dc4ad05f90564b0bb6dfc2a0e0da3cd4741254b4ccb0f8d267ebf14d6")

var userDB *sql.DB
var bookingDB *sql.DB

type Vehicle struct {
	ID   int
	Name string
}

type Booking struct {
	ID        int
	Vehicle   string
	StartTime string
	EndTime   string
	Price     float64
}

func InitBookingService() {
	var err error
	// Connect to the booking database for vehicles and bookings
	bookingDB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/booking_db")
	if err != nil {
		log.Fatalf("Failed to connect to booking database: %v", err)
	}
	if err := bookingDB.Ping(); err != nil {
		log.Fatalf("Booking database connection is not active: %v", err)
	}

	// Connect to the users database for user information
	userDB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/users_db")
	if err != nil {
		log.Fatalf("Failed to connect to users database: %v", err)
	}
	if err := userDB.Ping(); err != nil {
		log.Fatalf("Users database connection is not active: %v", err)
	}
}

func renderTemplate(w http.ResponseWriter, tmpl string, data interface{}) {
	t, err := template.ParseFiles(tmpl)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	t.Execute(w, data)
}

// View available vehicles
func ViewVehiclesHandler(w http.ResponseWriter, r *http.Request) {
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

	// Fetch available vehicles from bookingDB
	rows, err := bookingDB.Query("SELECT id, name FROM vehicles WHERE is_available = TRUE")
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

// Book a vehicle
func BookVehicleHandler(w http.ResponseWriter, r *http.Request) {
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

		// Fetch user info from usersDB
		var userID int
		var membership string
		err = userDB.QueryRow("SELECT id, membership FROM users WHERE email = ?", email).Scan(&userID, &membership)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		vehicleID := r.FormValue("vehicle_id")
		startTime := r.FormValue("start_time")
		endTime := r.FormValue("end_time")

		// Calculate price based on membership using billingMS
		price := billingMS.CalculatePrice(membership, startTime, endTime)

		// Insert booking into bookingDB
		_, err = bookingDB.Exec(
			"INSERT INTO bookings (user_id, vehicle_id, start_time, end_time, price) VALUES (?, ?, ?, ?, ?)",
			userID, vehicleID, startTime, endTime, price,
		)
		if err != nil {
			http.Error(w, "Error creating booking", http.StatusInternalServerError)
			return
		}

		// Mark vehicle as unavailable in bookingDB
		_, err = bookingDB.Exec("UPDATE vehicles SET is_available = FALSE WHERE id = ?", vehicleID)
		if err != nil {
			http.Error(w, "Error updating vehicle availability", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
	}
}

// View user bookings
func ViewBookingsHandler(w http.ResponseWriter, r *http.Request) {
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

	// Fetch user ID from usersDB
	var userID int
	err = userDB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusInternalServerError)
		return
	}

	// Fetch user bookings from bookingDB
	rows, err := bookingDB.Query(`
		SELECT b.id, v.name, b.start_time, b.end_time 
		FROM bookings b 
		JOIN vehicles v ON b.vehicle_id = v.id 
		WHERE b.user_id = ?`, userID)
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
func CancelBookingHandler(w http.ResponseWriter, r *http.Request) {
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

		// Fetch user ID from usersDB
		var userID int
		err = userDB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
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

		// Get vehicle ID for the booking from bookingDB
		var vehicleID int
		err = bookingDB.QueryRow("SELECT vehicle_id FROM bookings WHERE id = ? AND user_id = ?", bookingID, userID).Scan(&vehicleID)
		if err != nil {
			http.Error(w, "Booking not found", http.StatusInternalServerError)
			return
		}

		// Delete booking from bookingDB
		_, err = bookingDB.Exec("DELETE FROM bookings WHERE id = ?", bookingID)
		if err != nil {
			http.Error(w, "Error cancelling booking", http.StatusInternalServerError)
			return
		}

		// Mark vehicle as available in bookingDB
		_, err = bookingDB.Exec("UPDATE vehicles SET is_available = TRUE WHERE id = ?", vehicleID)
		if err != nil {
			http.Error(w, "Error updating vehicle availability", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, "/bookings", http.StatusSeeOther)
	}
}

// Update a booking
func UpdateBookingHandler(w http.ResponseWriter, r *http.Request) {
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
		err = userDB.QueryRow("SELECT id FROM users WHERE email = ?", email).Scan(&userID)
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
		_, err = bookingDB.Exec(
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
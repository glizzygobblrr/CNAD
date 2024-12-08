package billingMS

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/dgrijalva/jwt-go"
)

var userDB *sql.DB
var bookingDB *sql.DB
var billingDB *sql.DB

// Initialize database connection
func InitBillingService() {
	var err error
	bookingDB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/booking_db")
	if err != nil {
		log.Fatalf("Failed to connect to booking database: %v", err)
	}
	if err := bookingDB.Ping(); err != nil {
		log.Fatalf("Booking database connection is not active: %v", err)
	}

	userDB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/users_db")
	if err != nil {
		log.Fatalf("Failed to connect to users database: %v", err)
	}
	if err := userDB.Ping(); err != nil {
		log.Fatalf("Users database connection is not active: %v", err)
	}

	billingDB, err = sql.Open("mysql", "user:password@tcp(127.0.0.1:3306)/billing_db")
	if err != nil {
		log.Fatalf("Failed to connect to users database: %v", err)
	}
	if err := billingDB.Ping(); err != nil {
		log.Fatalf("Users database connection is not active: %v", err)
	}
}

// Calculate the price based on membership
func CalculatePrice(membership string, startTime, endTime string) float64 {
	duration := calculateDuration(startTime, endTime)
	basePrice := 10.0 // Base price per hour

	if membership == "Premium" {
		return basePrice * 0.9 * duration
	}
	return basePrice * duration 
}

// Calculate rental duration in hours
func calculateDuration(startTime, endTime string) float64 {
	start, _ := time.Parse("2006-01-02 15:04:05", startTime)
	end, _ := time.Parse("2006-01-02 15:04:05", endTime)
	return end.Sub(start).Hours() // Duration in hours
}

// Generate invoice
func GenerateInvoice(userID int, vehicleID int, startTime, endTime string, price float64) {
	invoiceID := fmt.Sprintf("INV-%d-%d", userID, time.Now().Unix())
	invoice := fmt.Sprintf("Invoice ID: %s\nUser ID: %d\nVehicle ID: %d\nStart Time: %s\nEnd Time: %s\nPrice: $%.2f",
		invoiceID, userID, vehicleID, startTime, endTime, price)

	
	log.Printf("Generated Invoice: \n%s\n", invoice)
	// Cannot implement

}

// Booking handler that calculates price and generates an invoice
func BookVehicleHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		// Get user authentication cookie
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Parse JWT claims
		claims := &jwt.MapClaims{}
		_, err = jwt.ParseWithClaims(cookie.Value, claims, func(token *jwt.Token) (interface{}, error) {
			return []byte("f0a567902063bf0d5d1e6c0c73a4666e70698882aa9b0ae508c7265efdc52865399728aee1a08bb17ef5de60fc8828ade449fde6f8ca1ed30244fef462c4f39f74c37dc80d1dea2873e7cee198d8a333ea5fe4b494c2f3bf9dc5e535cd8812442a0b6b4d7a93d47fb428e3320ec8448c314c576fabbbadb593489299b877d2a5557d1ef5dca07231b02f01c0a0ba8b043975ed38e81f736761b6e3db3a54847da5f1f29ec31426c0aa10308ea1ba35575b4b936d3cc6903e86afa33a3539ce9b437b6d22a04359061276a8a4cc9054c14a8db10625f07ab9d1ce345d91b717519923887dc4ad05f90564b0bb6dfc2a0e0da3cd4741254b4ccb0f8d267ebf14d6"), nil
		})
		if err != nil {
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		// Get user information
		email := (*claims)["email"].(string)
		var userID int
		var membership string
		err = userDB.QueryRow("SELECT id, membership FROM users WHERE email = ?", email).Scan(&userID, &membership)
		if err != nil {
			http.Error(w, "User not found", http.StatusInternalServerError)
			return
		}

		// Retrieve vehicle and rental details
		vehicleIDStr := r.FormValue("vehicle_id")
		vehicleID, err := strconv.Atoi(vehicleIDStr)
		if err != nil {
			http.Error(w, "Invalid vehicle ID", http.StatusBadRequest)
			return
		}
		startTime := r.FormValue("start_time")
		endTime := r.FormValue("end_time")

		// Calculate the price based on the user's membership
		price := CalculatePrice(membership, startTime, endTime)

		// Insert booking into the database
		_, err = bookingDB.Exec(
			"INSERT INTO bookings (user_id, vehicle_id, start_time, end_time, price) VALUES (?, ?, ?, ?, ?)",
			userID, vehicleID, startTime, endTime, price,
		)
		if err != nil {
			http.Error(w, "Error creating booking", http.StatusInternalServerError)
			return
		}

		_, err = bookingDB.Exec("UPDATE vehicles SET is_available = FALSE WHERE id = ?", vehicleID)
		if err != nil {
			http.Error(w, "Error updating vehicle availability", http.StatusInternalServerError)
			return
		}

		// Generate the invoice after booking
		GenerateInvoice(userID, vehicleID, startTime, endTime, price)

		http.Redirect(w, r, "/vehicles", http.StatusSeeOther)
	}
}

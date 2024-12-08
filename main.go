package main

import (
	"fmt"
	"log"
	"net/http"

	"CNAD/usersMS"
	"CNAD/bookingMS"
	//"CNAD/billingMS"
)


func main() {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	})
	
	// Initialize the user and booking services
	usersMS.InitUserService()
	bookingMS.InitBookingService()

	// Define routes for user management and vehicle reservations
	http.HandleFunc("/register", usersMS.RegisterHandler)
	http.HandleFunc("/login", usersMS.LoginHandler)
	http.HandleFunc("/update_user", usersMS.UpdateUserHandler)

	http.HandleFunc("/vehicles", bookingMS.ViewVehiclesHandler)
	http.HandleFunc("/book", bookingMS.BookVehicleHandler)
	http.HandleFunc("/bookings", bookingMS.ViewBookingsHandler)
	http.HandleFunc("/cancel_booking", bookingMS.CancelBookingHandler)
	http.HandleFunc("/update_booking", bookingMS.UpdateBookingHandler)


	// Start the server
	fmt.Println("Server started at http://localhost:5000")
	log.Fatal(http.ListenAndServe(":5000", nil))
}

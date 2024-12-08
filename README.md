Architecture diagram:
![Screenshot 2024-12-08 230953](https://github.com/user-attachments/assets/617801e1-9652-43d8-98e1-9c09db587374)

Steps to run:
1. Download the files
2. Set up the databases
3. Configure the databases (if you have changed the db name)
4. Run the program (go run main.go)
5. Access localhost://5000/login
6. Navigate through the application with the buttons

Design Consideration:

Separation of Services: Each microservice focuses on a specific domain (e.g., usersMS handles user-related actions,
bookingMS manages vehicle reservations, and billingMS deals with invoice generation). This ensures that each service 
is responsible for its own business logic and database management, reducing complexity and allowing for easier 
updates and maintenance.

Scalability: Each service can scale independently based on demand. For example, if there’s a high number of booking
requests, only the bookingMS can be scaled up without affecting the usersMS or billingMS. This allows efficient 
resource allocation and prevents bottlenecks in the system.

Database Per Service: Each microservice has its own database (e.g., usersMS uses a user database, bookingMS uses a 
booking-related database), preventing tightly coupled dependencies. This allows for independent scaling of the 
database and prevents performance degradation due to resource contention between services.

Inter-Service Communication: Microservices communicate via HTTP APIs as HTTP APIs are best suited for scenarios 
requiring immediate responses and real-time communication. Using HTTP makes it easier to implement JWT tokens for
enhanced security, and with standard HTTP methods like GET, POST, PUT, and DELETE, services can perform typical 
CRUD operations (Create, Read, Update, Delete) over the web. This simplifies how my microservices can handle various 
business logic tasks, such as user registration, login, booking vehicles.

Authentication & Authorization: Since the microservices are loosely coupled, security must be ensured at the service
boundary. For this, JWT tokens are used to authenticate users across services. Each service verifies the JWT token 
before allowing access to its endpoints.
Error Handling & Resilience: Proper error handling and fallbacks are implemented to ensure the system can handle failures gracefully. For instance, if one service is temporarily unavailable, users may receive informative error messages, and retry mechanisms are employed to maintain a robust system.

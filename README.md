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

1. Separation of Services: Each microservice focuses on a specific domain (e.g., usersMS handles user-related actions,
bookingMS manages vehicle reservations). This ensures that each service 
is responsible for its own business logic and database management, reducing complexity and allowing for easier 
updates and maintenance.

-- Benefits of separated services:
-> Domain-Driven Design: Ensures that each service has well-defined boundaries and responsibilities. This minimizes 
interdependencies and makes the system more modular.
-> Isolation of Failures: Issues in one service (e.g., billingMS) do not impact others (usersMS, bookingMS), reducing 
the risk of system-wide outages.Independent Evolution: Services can be developed, deployed, and updated independently 
without affecting the overall application.

2. Scalability: Each service (usersMS, bookingMS, billingMS) operates independently, allowing them to scale individually
based on their specific workload. For example, if there’s a high number of booking requests, only the bookingMS can be
scaled up without affecting the usersMS. This allows efficient resource allocation and prevents bottlenecks in the system.
Each microservice (e.g., usersMS) is designed to be stateless as they do not retain session information between requests.
All necessary data is included in each request (e.g., through JWTs for authentication) The server does not store session
information. Each request includes all necessary context (e.g., JWT tokens for authentication). This makes it easier to
scale horizontally since no dependency exists on server memory for session data.

--Benefits of a scalable application:
-> Efficient Resource Allocation: You can allocate resources where they are most needed. For instance: If there is a 
surge in user registrations, scale usersMS without allocating extra resources to bookingMS. During peak travel seasons,
scale bookingMS to handle increased booking traffic while leaving usersMS unaffected.
-> Statelessness: Each request includes all necessary context (e.g., JWT tokens for authentication). This makes it easier
to scale horizontally since no dependency exists on server memory for session data.

3. Database Per Service: Each microservice has its own database (e.g., usersMS uses a user database, bookingMS uses a 
booking-related database), preventing tightly coupled dependencies. This allows for independent scaling of the 
database and prevents performance degradation due to resource contention between services. Use relational databases
(MySQL) for services such as userMS and bookingMS, which require ACID compliance (atomicity, consistency, isolation,
and durability).

--Benefits of Database Per Service:
-> Scalability: Horizontal scaling becomes easier because each service can be scaled independently based on its specific 
workload.

4. Inter-Service Communication: Microservices communicate via HTTP APIs as HTTP APIs are best suited for scenarios 
requiring immediate responses and real-time communication. Using HTTP makes it easier to implement JWT tokens for
enhanced security, and with standard HTTP methods like GET, POST, PUT, and DELETE, services can perform typical 
CRUD operations (Create, Read, Update, Delete) over the web. This simplifies how my microservices can handle various 
business logic tasks, such as user registration, login, booking vehicles.

5. Authentication & Authorization: Since the microservices are loosely coupled, security must be ensured at the service
boundary. For this, JWT tokens are used to authenticate users across services. Each service verifies the JWT token 
before allowing access to its endpoints.
Error Handling & Resilience: Proper error handling and fallbacks are implemented to ensure the system can handle 
failures gracefully. MeANingful error logs and messages are also clearly stated to refer to in case of errors and bugs 
that requrie troubleshooting.

--Benefits of Authentication and Authorization: 
-> Security and Data Integrity: JWT tokens are signed with a secret key, ensuring that the contents of the token have 
not been tampered with during transmission. The token can also be encrypted to protect sensitive user data.
The JWT structure allows each service to independently verify the authenticity of the token, ensuring that only 
authorized users can access its resources.


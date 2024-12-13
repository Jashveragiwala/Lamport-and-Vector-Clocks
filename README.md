<div align="center" style="font-size: 20px;">

### Assignment 1

**Jash Jignesh Veragiwala (1006185)**

</div>

---

## Q1

This Go program simulates a client-server architecture using goroutines. In this simulation, several clients are registered to a server. Each client periodically sends a message to the server, and the server flips a coin to decide whether to forward the message to other clients or drop it. This implementation uses Go’s concurrency features to handle asynchronous messaging between clients and the server.

## Features

- **Client-Server Communication**: Clients periodically send messages to the server.
- **Message Broadcasting**: Upon receiving a message, the server either forwards it to other clients or drops it based on a random coin flip.
- **Concurrency**: The program uses goroutines to handle multiple clients and the server concurrently.

## Requirements

- **Go**: Ensure that Go is installed (version 1.13+ recommended).
- No external packages are used for this program.

## Part 1

In this parts its just basic implementtion of the above explained architecture.

## Program Structure

The program is structured around the **Client**, **Server**, and **Message** components. Each component includes specific functions and goroutines to handle message-sending, receiving, and broadcasting functionalities.

### Client Structure

1. **clientSender (goroutine)**:

   - Responsible for periodically sending messages from the client to the server. Each client instance sends a message every 500ms until it reaches the total message count set by the user. If the count is set to `-1`, it will send messages indefinitely. After sending all messages, it signals completion by sending a message to `closeChannel`.

2. **clientListener (goroutine)**:
   - Listens for messages broadcast by the server. When it receives a message, it displays the message information. Upon completion, it listens for a signal from `closeChannel` to stop listening for further messages and closes the client’s channel.

### Server Structure

1. **serverListener (goroutine)**:

   - Main listening function for the server, responsible for receiving messages from clients through `serverChannel`. When a message arrives, the server “flips a coin” to decide if the message should be forwarded or dropped. If forwarding, it calls `serverSender` to relay the message to all other clients. When all clients have completed sending messages, `serverListener` closes `serverChannel` to terminate message handling.

2. **serverSender (function)**:
   - Called by `serverListener` to broadcast a message to each client (except the original sender). This function directly relays a message from the server to the client’s `clientChannel`, allowing asynchronous message broadcasting.

### Message Structure

- **Message Struct**:
  - Represents each message sent from a client, containing a `senderID` (the client who sent it) and a `messageID` (a unique identifier for each message). This struct helps track each message's origin and identify it within the system.

### Main Program Flow

1. **main (function)**:
   - Initializes the server and client instances based on user input for the number of clients and messages. It sets up the server’s listening goroutine and each client’s sending and listening goroutines, managing concurrency through a wait group (`sync.WaitGroup`). After all goroutines complete their operations, the program exits with a final message.

## User Input and Output Explanation

When the program runs, it prompts the user to enter the number of clients and the number of messages each client will send. Here’s how it handles different input scenarios:

### Number of Clients:

The program requires at least 2 clients to operate. If the user inputs a number less than 2, it will display a message:

```
You need at least 2 clients to communicate.
```

### Number of Messages:

The user is prompted to enter the number of messages each client will send. The following scenarios are handled:

- If the user enters 0, it displays:

```
There is nothing to communicate. Please enter a valid number of messages.
```

- If the user enters -1, it indicates that the clients will send messages indefinitely.

## Output Color Codes

The program uses color codes to distinguish different types of output:

- **Green**: When the server receives a message.

  - Example:

  ```
  Server received Message 1 from Client 1
  ```

- **Blue**: When a client sends a message to the server.

  - Example:

  ```
  Client 1 is sending Message 1
  ```

- **Red**: When the server drops a message.

  - Example:

  ```
  Server has dropped Message 1 from Client 1
  ```

- **White**: For normal messages, such as client status updates.

## Final Output Messages

- The message:

```
All messages processed, program exiting.
```

indicates that the server has finished processing all messages from the clients and is shutting down.

- The message:

```
Client X is done listening for Messages...
```

indicates that a specific client has finished listening for messages and has exited the message-receiving loop.

## Compilation and Execution

To run the program, follow these steps:

1. **Change Directory**:

   ```bash
   cd Q1/Q1_1
   ```

2. **Running the Go file**:

   ```bash
   go run Q1_1.go
   ```

## Part 2

In this part, the program has been enhanced to implement **Lamport's logical clock** to determine a total order of all messages received at the registered clients. Most of the features are similar to Part 1, with minor modifications to incorporate the logical clock functionality:

- **Lamport Clock**: Each client now maintains a Lamport clock (`lamportClock`), which is updated on sending and receiving messages. The server also has its own Lamport clock to synchronize messages.

### Changes Made

1. **Message Struct**: Extended to include a `clock` field for the Lamport clock value.
2. **Client Sender**: Each client increments its Lamport clock before sending a message to the server.
3. **Server Listener**: The server updates its Lamport clock when receiving messages and uses it to update the clients' clocks before broadcasting messages.

During the execution of the program, as messages are transferred between the clients and the server, the Lamport clock values for both the server and each client are printed to the console. This provides real-time visibility into the logical timestamps associated with each message, illustrating the order of events in the distributed system. By displaying these clock values, users can gain insights into the synchronization of operations across different nodes, helping to understand how messages are processed and forwarded throughout the network.

## Compilation and Execution

To run the program, follow these steps:

1. **Change Directory**:

   ```bash
   cd Q1/Q1_2
   ```

2. **Running the Go file**:

   ```bash
   go run Q1_2.go
   ```

## Part 3

In this part, the program has been enhanced to implement **Vector Clocks** to manage concurrency and detect causality violations in message exchanges. Each client now maintains a vector clock that is updated upon sending and receiving messages. The server also maintains a vector clock for synchronization.

### Features Added

- **Vector Clock**: Each client has a `vectorTimeStamp` that tracks the logical time for each client.
- **Causality Detection**: The server checks the vector timestamp of each incoming message against its own. If the incoming message has a vector timestamp that indicates a causality violation, it prints a notification to the console.

### New Goroutines and Functions

1. **mergeVectorClock (function)**:

   - Merges two vector timestamps and increments the timestamp for the receiver. This function helps maintain a consistent view of the vector clock across clients and the server.

2. **client.prepMsgs (function)**:

   - Responsible for preparing and sending messages from the client to the server. It sends either a specified number of messages or enters an infinite loop to send messages indefinitely.

3. **server.serverListener (goroutine)**:

   - Listens for incoming messages from clients. It checks for causality violations and merges the vector clocks upon receiving messages. The server decides whether to forward the message to other clients or drop it based on a random choice.

4. **server.serverSender (function)**:

   - Broadcasts a received message to all clients except the original sender. It ensures that the message includes the updated vector clock.

5. **lesserVectorClock (function)**:
   - Compares two vector clocks to determine if the client’s clock indicates that it was generated before the server’s clock, aiding in causality detection.

### Changes Made

- **Message Struct**: Extended to include a `vectorTimeStamp` field for the vector clock values.
- **Client Sender**: Clients increment their vector clock before sending a message to the server.
- **Server Listener**: The server compares the vector clock of incoming messages to detect causality violations. If a violation is detected, it outputs a message highlighting the inconsistency.

During execution, the program will print detected causality violations, helping users understand the complexities of concurrency in distributed systems. Moreover, as messages are transferred between the clients and the server, the Lamport clock values for both the server and each client are printed to the console.

### Compilation and Execution

To run the program, follow these steps:

1. **Change Directory**:

   ```bash
   cd Q1/Q1_3
   ```

2. **Running the Go file**:

   ```bash
   go run Q1_3.go
   ```

## Q2

This Go program implements the Ring Protocol for replica synchronization in a distributed system. Each replica maintains a local data structure(in this case it is an integer) that can diverge for various reasons. The coordinator periodically sends its data to all other replicas, which update their local versions with the coordinator's data. In case of a coordinator failure, the program initiates a new election using the Ring algorithm to choose a new coordinator among the active processes. This simulation utilizes Go’s concurrency features to handle multiple processes running concurrently.

## Features

- **Replica Synchronization**: The coordinator sends data to all active processes, allowing them to update their local data structures.
- **Coordinator Election**: If the current coordinator crashes, the system detects the failure and elects a new coordinator using the Ring algorithm.
- **Concurrency**: The program utilizes goroutines to simulate multiple processes running in parallel, enhancing the responsiveness of the system.

## Requirements

- **Go**: Ensure that Go is installed (version 1.13+ recommended).
- No external packages are used for this program.

## Part 1

This program consists of multiple processes that communicate in a ring topology. Each process can be in an active or crashed state, and the coordinator is responsible for synchronizing data with the other processes. The election of a new coordinator is triggered when the current coordinator fails.

## Program Structure

The program is structured around the **Process** component, which includes methods for running the process, sending data, checking the coordinator's status, and handling elections.

### Process Structure

1. **run (goroutine)**:

   - Each process runs this function to manage its status. If it is the coordinator, it periodically sends data to other processes. If it is not the coordinator, it waits for data from the coordinator and checks its status.

2. **sendDataToProcesses**:

   - The coordinator sends its data to all other active processes, allowing them to update their local data structures.

3. **checkCoordinatorStatus**:

   - Each process checks if the coordinator has crashed. If so, it initiates an election to choose a new coordinator.

4. **initiateElection**:

   - This function starts the election process by passing a ring of active process IDs to the next process in the ring.

5. **receiveRing**:

   - Processes receive the ring of IDs and either add their own ID or complete the election if they find themselves already in the ring.

6. **updateRing**:

   - The new ring structure is updated, and the process with the highest ID is elected as the new coordinator.

7. **sendRingToNextActiveProcess**:

   - Passes the ring to the next active process. It searches for the next active process in the ring and sends the ring to it. If no active process is found, it logs that the election could not proceed.

8. **updateData**:

   - Updates the data for the process with the provided new data value. It locks the process to ensure thread safety during the update.

9. **findProcessByID**

   - Finds and returns a pointer to a process by its ID. If no process with the given ID is found, it returns nil.

10. **getActiveProcesses**

    - Returns a slice of pointers to all active processes (status = 1).

11. **crashProcess**

    - Simulates crashing a process by changing its status to inactive (0). It locks the process during this operation to ensure thread safety.

12. **randomlyChangeData**

    - Randomly changes data for non-coordinator active processes at random intervals. It ensures that the coordinator does not get selected for data updates.

13. **allProcessesCrashed**
    - Checks if all processes have crashed (status = 0). Returns true if all processes are inactive; otherwise, returns false.

### Main Program Flow

1. **Seed Random Number Generator**: The random number generator is seeded using the current time to ensure different random values on each execution.

2. **User Input**: The program prompts the user to enter the number of processes to be created in the system.

   ```go
   Enter the number of processes:
   ```

3. **Initialize Processes**: The specified number of processes is created. Each process is assigned a unique ID, an initial status of 1 (active), and a random data value between 0 and 100.

4. **Create Ring Structure**: Each process is linked in a ring structure by filling its ring array with the IDs of all processes in a circular manner.

5. **Select Initial Coordinator**: The process with the highest ID is designated as the initial coordinator.

6. **Start Processes**: Each process is started in a separate goroutine to allow concurrent execution.

7. **Simulate Random Data Changes**: A separate goroutine is launched to change the data of non-coordinator processes at random intervals.

8. **Randomly Crash and Activate Processes**: Another goroutine continuously checks if all processes have crashed. If not, it randomly selects a process to crash every 10 seconds.

9. **Wait for Completion**: The main function waits for all processes to finish before exiting.

## User Input and Output Explanation

- **Input**: The user is prompted to input the number of processes they wish to create in the system.

- **Output**:
  - After the user provides input, the program initializes the processes and prints the ID of the initial coordinator along with its ring structure.
  ```go
  Process <id> is the initial Coordinator with the following ring structure <ring>.
  ```
  - During execution, various activities are logged, including:
    - Passing the ring between processes
    - Updating data for processes
    - Crashing processes
    - Election completion and new coordinator announcements

## Output Color Codes

Output messages are color-coded to improve visibility and distinguish between types of events:

- **Green**: Information about processes passing the ring and updating the structure.
- **Yellow**: Data changes made to non-coordinator processes.
- **Blue**: Coordinator election announcements.
- **Red**: Crash notifications or termination messages.

## Running the Program

1. Navigate to the project directory.
   ```bash
   cd Q2/Q2_1
   ```
2. Run the program using the command:
   ```bash
   go run Q2_1.go
   ```

## Part 2

In this part, the program simulates both the **worst-case** and **best-case** scenarios of the ring election algorithm, where all machines start an election versus only one machine starting the election.

### Worst-Case Scenario

To demonstrate a **worst-case scenario** where multiple processes initiate the election simultaneously, the following changes were made to the code from previous part:

1. **Removed the `electionInProgress` Boolean**: This allows processes to initiate elections even if an election is already underway, simulating multiple concurrent elections.

2. **Coordinator Crash**: The `main` function is modified so that the initial coordinator is forced to leave the ring immediately, prompting multiple processes to start elections. This change speeds up the simulation of the worst-case scenario.

### Best-Case Scenario

The best-case scenario, where only one machine initiates the election, uses the same code as **Part 1**. Here, only a single process starts the election, avoiding simultaneous elections and providing a more straightforward coordination flow.

## Compilation and Execution

To run the program, follow these steps:

1. Navigate to the project directory.
   ```bash
   cd Q2/Q2_2
   ```
2. Run the program using the command:
   ```bash
   go run Q2_2.go
   ```

## Part 3

In this part, the program simulates 2 cases where a coordinator and a non coordinator fails during the election is going on.

### Part A

The following changes were made to the code from previous part:

1. **`electionInProgress` Boolean function is brought back**: This allows processes to initiate election only one at a time.

2. **Change in the `updateRing` function**: The **crashProcess** function is called within this function to crash the newly elected coordinator while the new ring structure is getting circulated within the ring during the annoucement stage.

### Part B

The following changes were made to the code from previous part:

1. **`electionInProgress` Boolean function is brought back**: This allows processes to initiate election only one at a time.

2. **Change in the `sendRingToNextActiveProcess` function**: The **crashProcess** function is called within this function to crash a random node that in not the new coordinator while the ring structure is getting circulated within the ring during the discovery stage.

## Compilation and Execution

To run the program, follow these steps:

1. Navigate to the project directory.

   ```bash
   cd Q2/Q2_3
   ```

To run **Part A**:

2. Run the program using the command:
   ```bash
   cd Q2_3A
   go run Q2_3A.go
   ```

<div align="center" style="font-weight:bold">OR</div>

To run **Part B**:

2. Run the program using the command:
   ```bash
   cd Q2_3B
   go run Q2_3B.go
   ```

## Part 4

In this part, the program simulates 2 cases where a coordinator or a non coordinator leave the ring silently. The outcome of this can be as follows:

1. If the **coordinator leaves** then an election is conducted and the ring structure gets updated for all the processes.
2. If a **non coordinator leaves** then nothing happens as the ring structure of the all the processes locally remains same until next election is conducted and new ring structure is updated.

These senarios are same as of **part 1**. Hence it can be run in similar fashion as part 1.

## Running the Program

1. Navigate to the project directory.
   ```bash
   cd Q2/Q2_1
   ```
2. Run the program using the command:
   ```bash
   go run Q2_1.go
   ```

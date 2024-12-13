package main

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"time"
)

type Process struct {
	id      int
	status  int // 1: active, 0: crashed
	data    int // local version of the data structure
	lock    sync.Mutex
	elected bool
	ring    []int
}

var processes []*Process
var wg sync.WaitGroup
var coordinator *Process
var electionInProgress = false // Flag to indicate if an election is in progress
var electionMutex sync.Mutex

func (p *Process) run() {
	defer wg.Done()
	for {
		if p.status == 1 {
			if p.id == coordinator.id {
				// The coordinator periodically sends its data to all processes every 4 seconds
				time.Sleep(4 * time.Second)
				if p.status == 1 { // Ensure the coordinator is still active
					p.sendDataToProcesses()
				}
			} else {
				// Check if it's been more than 4 seconds since the last data was received
				time.Sleep(4 * time.Second)
				p.checkCoordinatorStatus()
			}
		}
	}
}

// Function for the coordinator to send data to all processes
func (p *Process) sendDataToProcesses() {
	p.lock.Lock()
	defer p.lock.Unlock()
	for _, proc := range processes {
		if proc.status == 1 && proc.id != p.id {
			fmt.Printf("Coordinator %d is sending data %d to Process %d.\n", p.id, p.data, proc.id)
			proc.lock.Lock()
			if proc.status == 1 {
				proc.data = p.data
				fmt.Printf("Process %d updated its data to: %d (received from Coordinator %d)\n", proc.id, proc.data, p.id)
			}
			proc.lock.Unlock()
		}
	}
}

// Function to check the status of the coordinator and initiate an election if necessary
func (p *Process) checkCoordinatorStatus() {
	if coordinator.status == 0 {
		electionMutex.Lock()
		if !electionInProgress {
			electionInProgress = true
			fmt.Printf("\033[32mProcess %d detects that Coordinator %d has crashed, initiating election.\033[0m\n", p.id, coordinator.id)
			go p.initiateElection() // Start election from this process

		}
		electionMutex.Unlock()

	}
}

// Function for a process to initiate an election
func (p *Process) initiateElection() {
	electionRing := []int{p.id}
	fmt.Printf("\033[32mProcess %d is starting the election, initial ring: %v\033[0m\n", p.id, electionRing)
	p.sendRingToNextActiveProcess(electionRing)
}

// Function to pass the ring to the next active process
func (p *Process) sendRingToNextActiveProcess(ring []int) {
	for i := 1; i < len(p.ring); i++ {
		nextProcessID := p.ring[i]
		nextProcess := findProcessByID(nextProcessID)
		if nextProcess != nil && nextProcess.status == 1 {
			fmt.Printf("\033[32mProcess %d passing ring %v to Process %d\033[0m\n", p.id, ring, nextProcess.id)
			nextProcess.receiveRing(ring)
			return
		}
	}
	// No active process found, end the election
	fmt.Printf("\033[32mProcess %d could not find any active process to pass the ring.\033[0m\n", p.id)
}

// Function to receive the ring and process it
func (p *Process) receiveRing(ring []int) {
	// Check if the current process's ID is already in the ring
	for _, id := range ring {
		if id == p.id {
			// Election complete, update the new ring structure
			fmt.Printf("\033[32mProcess %d found its ID in the ring. Updating new ring structure: %v\033[0m\n", p.id, ring)
			p.updateRing(ring)
			return
		}
	}

	// Add the current process's ID to the ring
	newRing := append(ring, p.id)
	fmt.Printf("\033[32mProcess %d adding itself to the ring, new ring: %v\033[0m\n", p.id, newRing)
	p.sendRingToNextActiveProcess(newRing)
}

// Function to update the ring structure for all active processes
func (p *Process) updateRing(newRing []int) {
	for i, id := range newRing {
		process := findProcessByID(id)
		if process != nil && process.status == 1 {
			modifiedRing := append(newRing[i:], newRing[:i]...)
			process.lock.Lock()
			process.ring = modifiedRing
			process.lock.Unlock()
			fmt.Printf("\033[32mProcess %d updated with new ring structure: %v\033[0m\n", process.id, modifiedRing)
		}
	}

	// Elect the process with the highest ID as the coordinator
	maxID := newRing[0]
	for _, id := range newRing {
		if id > maxID {
			maxID = id
		}
	}
	newCoordinator := findProcessByID(maxID)
	if newCoordinator != nil {
		newCoordinator.elected = true
		coordinator = newCoordinator
		fmt.Printf("\033[34mProcess %d is elected as the new Coordinator.\033[0m\n", newCoordinator.id)
	}
	electionInProgress = false
}

// Method to update the data for a process
func (p *Process) updateData(newData int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.data = newData
	fmt.Printf("\033[33mProcess %d changed its data to: %d\033[0m\n", p.id, newData)
}

// Function to find a process by its ID
func findProcessByID(id int) *Process {
	for _, proc := range processes {
		if proc.id == id {
			return proc
		}
	}
	return nil
}

// Function to get the list of active processes
func getActiveProcesses() []*Process {
	active := []*Process{}
	for _, proc := range processes {
		if proc.status == 1 {
			active = append(active, proc)
		}
	}
	return active
}

// Function to crash a process
func crashProcess(id int) {
	for _, proc := range processes {
		if proc.id == id && proc.status == 1 {
			proc.lock.Lock()
			proc.status = 0
			proc.lock.Unlock()
			fmt.Printf("\033[31mProcess %d crashed (This process leaves silently, its not annoucement. Just for us to know when a process has crashed.).\033[0m\n", id)
			break
		}
	}
}

// Function to randomly change data for non-coordinator active processes
func randomlyChangeData() {
	for {
		time.Sleep(time.Duration(rand.Intn(10)+5) * time.Second)
		activeProcesses := getActiveProcesses()

		// Exclude the coordinator from the selection
		if len(activeProcesses) <= 1 {
			continue
		}

		var targetProcess *Process
		for {
			targetProcess = activeProcesses[rand.Intn(len(activeProcesses))]
			if targetProcess != coordinator {
				break
			}
		}

		// Update the data of the selected process
		newData := rand.Intn(100)
		fmt.Printf("\033[33mChanging data for Process %d to: %d\033[0m\n", targetProcess.id, newData)
		targetProcess.updateData(newData)
	}
}

func allProcessesCrashed() bool {
	for _, proc := range processes {
		if proc.status == 1 {
			return false
		}
	}
	return true
}

func main() {
	rand.Seed(time.Now().UnixNano())
	var numProcesses int
	fmt.Print("\033[38;5;88mEnter the number of processes: \033[0m")
	fmt.Scanln(&numProcesses)
	// Initialize processes and create the ring structure
	for i := 1; i <= numProcesses; i++ {
		processes = append(processes, &Process{id: i, status: 1, data: rand.Intn(100)})
	}

	// Link the process IDs in the ring
	for i := 0; i < numProcesses; i++ {
		processes[i].ring = make([]int, numProcesses)
		for j := 0; j < numProcesses; j++ {
			processes[i].ring[j] = processes[(i+j)%numProcesses].id // Fill ring with IDs in order
		}
	}

	// Set the initial coordinator to the process with the highest ID
	coordinator = processes[0]
	for _, proc := range processes {
		if proc.id > coordinator.id {
			coordinator = proc
		}
	}
	coordinator.elected = true // Mark as the coordinator
	fmt.Printf("\033[34mProcess %d is the initial Coordinator with the following ring structure %v.\033[0m\n", coordinator.id, coordinator.ring)
	for _, proc := range processes {
		wg.Add(1)
		go proc.run()
	}
	go randomlyChangeData()

	// Randomly crash and activate processes
	go func() {
		for {
			if allProcessesCrashed() {
				fmt.Println("\033[31mAll processes have ended. Terminating program.\033[0m")
				os.Exit(0)
			}
			time.Sleep(10 * time.Second)
			crashProcess(rand.Intn(numProcesses) + 1) // Crash a random process
			time.Sleep(10 * time.Second)
		}
	}()
	wg.Wait()
}

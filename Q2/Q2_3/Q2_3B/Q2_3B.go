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
var electionInProgress = false
var electionMutex sync.Mutex
var crashedDuringElection = false // Flag to ensure only one non-coordinator crashes during an election

func (p *Process) run() {
	defer wg.Done()
	for {
		if p.status == 1 {
			if p.id == coordinator.id {
				time.Sleep(4 * time.Second)
				if p.status == 1 {
					p.sendDataToProcesses()
				}
			} else {
				time.Sleep(4 * time.Second)
				p.checkCoordinatorStatus()
			}
		}
	}
}

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

func (p *Process) checkCoordinatorStatus() {
	if coordinator.status == 0 {
		electionMutex.Lock()
		if !electionInProgress {
			electionInProgress = true
			fmt.Printf("\033[32mProcess %d detects that Coordinator %d has crashed, initiating election.\033[0m\n", p.id, coordinator.id)
			go p.initiateElection()
		}
		electionMutex.Unlock()
	}
}

func (p *Process) initiateElection() {
	electionRing := []int{p.id}
	fmt.Printf("\033[32mProcess %d is starting the election, initial ring: %v\033[0m\n", p.id, electionRing)
	p.sendRingToNextActiveProcess(electionRing)
}

func (p *Process) sendRingToNextActiveProcess(ring []int) {
	for i := 1; i < len(p.ring); i++ {
		nextProcessID := p.ring[i]
		nextProcess := findProcessByID(nextProcessID)
		if nextProcess != nil && nextProcess.status == 1 {
			// Introduce a forced crash of a random non-coordinator process during election
			if !crashedDuringElection && nextProcess.id != coordinator.id && rand.Intn(2) == 0 {
				crashProcess(nextProcess.id)
				crashedDuringElection = true
				fmt.Printf("\033[31mNon-coordinator Process %d crashed during election.\033[0m\n", nextProcess.id)
				continue
			}

			fmt.Printf("\033[32mProcess %d passing ring %v to Process %d\033[0m\n", p.id, ring, nextProcess.id)
			nextProcess.receiveRing(ring)
			return
		}
	}
	fmt.Printf("\033[32mProcess %d could not find any active process to pass the ring.\033[0m\n", p.id)
}

func (p *Process) receiveRing(ring []int) {
	for _, id := range ring {
		if id == p.id {
			fmt.Printf("\033[32mProcess %d found its ID in the ring. Updating new ring structure: %v\033[0m\n", p.id, ring)
			p.updateRing(ring)
			return
		}
	}

	newRing := append(ring, p.id)
	fmt.Printf("\033[32mProcess %d adding itself to the ring, new ring: %v\033[0m\n", p.id, newRing)
	p.sendRingToNextActiveProcess(newRing)
}

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
	crashedDuringElection = false // Reset for next election
}

func (p *Process) updateData(newData int) {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.data = newData
	fmt.Printf("\033[33mProcess %d changed its data to: %d\033[0m\n", p.id, newData)
}

func findProcessByID(id int) *Process {
	for _, proc := range processes {
		if proc.id == id {
			return proc
		}
	}
	return nil
}

func crashProcess(id int) {
	for _, proc := range processes {
		if proc.id == id && proc.status == 1 {
			proc.lock.Lock()
			proc.status = 0
			proc.lock.Unlock()
			fmt.Printf("\033[31mProcess %d crashed.\033[0m\n", id)
			break
		}
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

	for i := 1; i <= numProcesses; i++ {
		processes = append(processes, &Process{id: i, status: 1, data: rand.Intn(100)})
	}

	for i := 0; i < numProcesses; i++ {
		processes[i].ring = make([]int, numProcesses)
		for j := 0; j < numProcesses; j++ {
			processes[i].ring[j] = processes[(i+j)%numProcesses].id
		}
	}

	coordinator = processes[0]
	for _, proc := range processes {
		if proc.id > coordinator.id {
			coordinator = proc
		}
	}
	coordinator.elected = true
	fmt.Printf("\033[34mProcess %d is the initial Coordinator with the following ring structure %v.\033[0m\n", coordinator.id, coordinator.ring)

	for _, proc := range processes {
		wg.Add(1)
		go proc.run()
	}

	go func() {
		time.Sleep(10 * time.Second)
		crashProcess(coordinator.id)
		fmt.Printf("\033[31mCoordinator Process %d crashed forcefully.\033[0m\n", coordinator.id)

		// Randomly and slowly crash remaining processes
		for {
			if allProcessesCrashed() {
				fmt.Println("\033[31mAll processes have ended. Terminating program.\033[0m")
				os.Exit(0)
			}
			time.Sleep(5 * time.Second)
			crashProcess(rand.Intn(numProcesses) + 1)

		}
	}()

	wg.Wait()
}

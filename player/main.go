package main

import (
	"fmt"
)

func main() {
	fmt.Println("OSCDP (Open Source CD Player)")
	fmt.Println("2024 - Danilo Fragoso")
	fmt.Println("--------------")

	controller, err := InitController()
	if err != nil {
		fmt.Printf("Failed to initialize controller: %v\n", err)
		fmt.Println("Continuing without controller support")
	}

	mpv, err := InitMPV()
	if err != nil {
		fmt.Printf("Failed to initialize MPV: %v\n", err)
		return
	}

	player := InitPlayer(mpv)

	discSize := make(chan int64)
	go monitorDiscSize(discSize)

	controllerKeyPresses := make(chan string)
	if controller != nil {
		fmt.Println("Controller initialized")
		go controller.ListenKeys(controllerKeyPresses)
	}

	for {
		select {
		case size := <-discSize:
			if player.Disc == nil || player.Disc.Size != size {
				var err error
				fmt.Println("Detecting new disc")
				player.Disc, err = createAndIdentifyDisk(size)
				if err != nil {
					player.EjectDisc()
					continue
				}

				fmt.Println("New disc detected")
				fmt.Println("Artist:", player.Disc.Artist)
				fmt.Println("Title:", player.Disc.Title)

				if err := player.StartDisc(); err != nil {
					player.EjectDisc()
				}
			}
		case key := <-controllerKeyPresses:
			player.HandleKey(key)
		}

		player.UpdatePosition()
		player.UpdateStatus()

		if controller != nil {
			player.UpdateController(controller)
		}
	}
}

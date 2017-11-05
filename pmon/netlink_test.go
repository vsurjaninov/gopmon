package pmon

import (
	"fmt"
	"time"
	"testing"
)

func TestAck(t *testing.T) {
	listener := &ProcListener{}
	acks := make([]EventAck, 1)

	err := listener.Connect()
	if err != nil {
		t.Fatal("Failed connect")
	}

	defer listener.Close()
	go listener.ListenEvents()
	time.Sleep(1)

	select {
	case err := <-listener.Error:
		fmt.Println("error on receive: ", err)
		t.Fatal("Error on recv")
	case event := <-listener.EventAck:
		fmt.Printf("%T no=%d\n", event, event.No)
		acks = append(acks, *event)
		listener.Close()
		return
	}

	t.Errorf("No events received")
}

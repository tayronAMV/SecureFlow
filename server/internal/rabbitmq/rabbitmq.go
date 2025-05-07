package rabbitmq

import (
	"fmt"
	"github.com/streadway/amqp"
)



type interval_log struct{
	contain
}

//TODO : change to the correct credentials 
func connect_to_agent() error {
	conn , err := amqp.Dial("amqp://guest:guest@localhost:5672/") 
	if err != nil{
		fmt.Println(err)
		panic(err)
	}

	defer conn.Close()

	ch , err := conn.Channel()

	if err != nil{
		fmt.Println(err)
		panic(err)
	}

	defer ch.Close()


	return nil 
}
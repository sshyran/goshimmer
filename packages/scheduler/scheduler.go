package scheduler

// inbox buffered channel
// list for managing message with timestamps in the future
// multi-map
// outbox buffered channel

// channel for the inbox
// Outside of the scheduker, you fetch and evaluate if it has booked parent, otherswise you skip it (also the timestamp)

// single background goroutine with a select:
// case1:	reads from the inbox channel:
//			- when something comes in, it puts it in the right position of the oredered list
//  		- looks at the parent and creates a multi-map (key: parentID, value: []message)
// case2: 	listen to timer for when the first element of the ordered list gets ready
// 			- check if more elements are also ready and update the timer accordingly to the first in the future.
//  		- the ready elements get removed and moved in the multi-map

// case3: 	listen to messageBooked channel
//			- check all the messages associated to the newly booked parents in the multi-map
//			- if yes, we put the message into the FIFO (outbox buffered channel)

// listening to the MessageBooked event:
//  - translated to a buffered channel with the messageID of the newly booked

// remember to handle the shutting down with closing the channels and so on...

// have a look at the matcher of the autopeering server.go

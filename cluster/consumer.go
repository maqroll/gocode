package main

import zmq "gozmq"
import "flag"
import "strconv"
import "fmt"
import "runtime"

var ip = flag.String("ip","127.0.0.1","ip")
var port = flag.Int("port",6000,"port")
var portEOF = flag.Int("portEOF",7000,"port to notify EOF")
var context zmq.Context

func main() {
    context, _ /*TODO*/ = zmq.NewContext()

    finished := make(chan bool)
    rcvEOF := make(chan bool)

    go loopRcv(rcvEOF, finished)

    go testEOF(rcvEOF)

    <-finished
    context.Close()
}

func loopRcv(rcvEOF chan bool, finished chan bool) {
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    socket, _ /*TODO*/ := context.NewSocket(zmq.PULL)
     socket.Connect("tcp://" + *ip + ":" + strconv.Itoa(*port))
    defer socket.Close()

    pi := zmq.PollItems{ zmq.PollItem{Socket: socket, Events: zmq.POLLIN} } 

    for {
        /*TODO*/ zmq.Poll(pi, 5000)

        if pi[0].REvents&zmq.POLLIN != 0 {
                line, _ := socket.Recv(0) 
                fmt.Printf("%s\n",line)
        } else {
            select {
                case <-rcvEOF:
                    finished <- true
                    return
                default:
            }
        }
    }
}

func testEOF(rcvEOF chan bool) {
    runtime.LockOSThread()
    defer runtime.UnlockOSThread()
    eof,_ /*TODO*/ := context.NewSocket(zmq.SUB)
    eof.Connect("tcp://" + *ip + ":" + strconv.Itoa(*portEOF))
    eof.SetSockOptString(zmq.SUBSCRIBE,"")

    eof.Recv(0)
    eof.Close()
    rcvEOF <- true
}

package main

import "os"
import "bufio"
import zmq "gozmq"
import "flag"
import "strconv"
//import "time"
import "runtime"

var port = flag.Int("port",6000,"port")
var portEOF = flag.Int("portEOF",7000,"port to notify EOF")

func main() {
    runtime.LockOSThread()
    context, _ /*TODO*/ := zmq.NewContext()
    socket, _ /*TODO*/ := context.NewSocket(zmq.PUSH)
    socket.Bind("tcp://*:" + strconv.Itoa(*port))
    contextEOF, _ /*TODO*/ := zmq.NewContext()
    eof,_ /*TODO*/ := contextEOF.NewSocket(zmq.PUB)
    eof.Bind("tcp://*:" + strconv.Itoa(*portEOF))

    socket.SetSockOptInt(zmq.LINGER,5000)
    eof.SetSockOptInt(zmq.LINGER,5000)

    var partialLine []byte = nil

    bufStdin := bufio.NewReader(os.Stdin)
    
    line, isPrefix, err := bufStdin.ReadLine()
    for err == nil {
        if isPrefix || partialLine!=nil {
            partialLine = append(partialLine,line...)
        } 

        if !isPrefix {
            if partialLine!=nil { 
                socket.Send(partialLine,0)
                partialLine = nil
            } else { 
                socket.Send(line,0)
            }
        }
        
        line, isPrefix, err = bufStdin.ReadLine()
    }
    socket.Close()
    context.Close()

    eof.Send([]byte("EOF"),0)

    eof.Close()
    contextEOF.Close()
}

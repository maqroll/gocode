package main

import (
    "os"
    "log"
    "bufio"
    "io"
    "flag"
    "strings"
    "fmt"
    "encoding/json"
)

var inputFlag = flag.String("in", "./in.txt", "input file")

var lines = make(chan string,100)
var end = make(chan int)

func process() {
    var prevLine *Line
    var stacks int
    var currentStack = make(map[string]*Line)
    var idxLines = make(map[string]*Line)
    var id = 0

    for line := range lines {
        if (strings.TrimSpace(line) == "") {
            stacks = stacks + 1
            prevLine = nil
            for k := range currentStack {
                delete(currentStack, k)
            }
            continue
        } else {
            line = line[strings.Index(line,"(") + 1 : len(line)-1]
        }

        var node *Line

        if (idxLines[line] != nil) {
            if (currentStack[line] != nil) {
                currentStack[line].Recursive = true
                //node = idxLines[line]
                // fmt.Println("Ignoring " + line)
                continue
            } else {
                node = idxLines[line]
                node.Occurrences++
            }
        } else {
            node = new(Line)
            node.Path = line
            node.Occurrences = 1
            node.Recursive = false
            node.Id = id
            id++
            node.Callers = make(map[string]*Call,10)
            node.Callees = make(map[string]*Line,10)

            idxLines[line] = node
        }

        currentStack[line] = node   
        if (prevLine != nil) {
            if (node.Callers[prevLine.Path] == nil) {
                call := new(Call)
                call.Occurrences = 1
                call.Line = prevLine
                node.Callers[prevLine.Path] = call
            } else {
                node.Callers[prevLine.Path].Occurrences++
            }
            prevLine.Callees[node.Path] = node
        }
        
        prevLine = node
    }

    var nodesLen = 0
    /* Recorremos idxLines y colapsamos los nodos */
    for nodesLen != len(idxLines) {
        nodesLen = len(idxLines)
        for _,n := range idxLines {
            //fmt.Println("Processing " + n.Path)
            if len(n.Callers) == 1 {
                collapsable := true
                var collapse *Call
                for _,a := range n.Callers {
                    if len(a.Line.Callees) != 1 {
                        collapsable = false
                    } else {
                        collapse = a
                    }
                } 

                if collapsable {
                    /* colapsamos el llamado en el llamante */
                    //fmt.Println("Collapsing " + collapse.Line.Path + " into " + n.Path)
                    delete(idxLines,n.Path)
                    delete(idxLines,collapse.Line.Path)
                    n.Path = n.Path + "|" + collapse.Line.Path
                    n.Callers = collapse.Line.Callers
                    for _,j := range n.Callers {
                        /* Recorremos todos los que nos llaman y actualizamos la referencia */
                        delete(j.Line.Callees,collapse.Line.Path)
                        j.Line.Callees[n.Path] = n
                    }
                    idxLines[n.Path] = n

                    // for _,k := range idxLines {
                    //     fmt.Println("->" + k.Path)
                    //     for _,d := range k.Callers {
                    //         fmt.Println("*" + d.Line.Path)
                    //     }
                    // }
                } 
            }
        }
    }

    /* Deberiamos volcar el contenido del fichero a json */
    sankey := new(Sankey)
    sankey.Nodes = make([]*Node, len(idxLines))
    sankey.Links = make([]*Link, 0, 50)

    i := 0
    for _,n := range idxLines {
        node := new(Node)
        node.Name = n.Path
        n.Id = i
        sankey.Nodes[i] = node
        i++
    }    

    /* Faltan los enlaces */
    for _,l := range idxLines {
        for _,c := range l.Callers {
            link := new(Link)
            link.Target = l.Id
            link.Source = c.Line.Id
            // link.Luis = c.Line.Path
            link.Value = c.Occurrences
            sankey.Links = append(sankey.Links,link)
        }
    }

    o,_ := json.Marshal(sankey)
    fmt.Println(string(o))

    end <- 1
}

func main() {
    flag.Parse()

    go process()

    f, err := os.Open(*inputFlag) 
    if err != nil {           
        log.Fatal(err)
    }
 
    bf := bufio.NewReader(f)

    stack := make([]string,0,50)
 
    for {
        bytes, isPrefix, err := bf.ReadLine()
 
        if err == io.EOF {
            for i := 1; i <= len(stack); i++ {
               lines <- stack[len(stack)-i]
            }       
            break
        }
 
        if err != nil {
            log.Fatal(err)
        }
 
        if isPrefix {
            log.Fatal("Error: Unexpected long line reading", f.Name())
        }
 
        line := string(bytes)

        if strings.TrimSpace(line) != "" {
            stack = append(stack,line)
        } else {
            /* enviarlos al reves */
            for i := 1; i <= len(stack); i++ {
               lines <- stack[len(stack)-i]
            }       
            lines <- ""
            stack = make([]string,0,50)
        }
    }
    close(lines)
    <- end
}
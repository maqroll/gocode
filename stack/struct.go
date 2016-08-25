package main

type Call struct {
    Line *Line
    Occurrences int
}

type Line struct {
    Id int
    Path string
    Occurrences int
    Recursive bool
    Callers map[string]*Call /* quien nos llama */
    Callees map[string]*Line /* a quien llamamos */
}

type Sankey struct {
    Nodes []*Node `json:"nodes"`
    Links []*Link `json:"links"`
}

type Node struct {
    Name string `json:"name"`
}

type Link struct {
    Source int `json:"source"`
    Target int `json:"target"`
    // Luis string 
    Value int `json:"value"`
}
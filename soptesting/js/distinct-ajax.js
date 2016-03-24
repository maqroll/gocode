// GET to different origin (with result)
// fails on the client side but the requests get executed
var getDistinct = new XMLHttpRequest();
getDistinct.open('GET', 'http://localhost:9091/json?value=ok', true);

getDistinct.onload = function() {
  if (getDistinct.status >= 200 && getDistinct.status < 400) {
    var data = JSON.parse(getDistinct.responseText);
    document.getElementById("get-distinct").innerHTML = data.key;
  } else {
    document.getElementById("get-distinct").innerHTML = "error";
  }
};

getDistinct.onerror = function() {
    document.getElementById("get-distinct").innerHTML = "error";
};

getDistinct.send();


// GET to different origin (without result)
// fails on the client side but the requests get executed
var getWODistinct = new XMLHttpRequest();
getWODistinct.open('GET', 'http://localhost:9091/json?value=ok', true);

getWODistinct.onload = function() {
//  if (getWODistinct.status >= 200 && getWODistinct.status < 400) {
    document.getElementById("get-wo-distinct").innerHTML = "ok";
//  } else {
//    document.getElementById("get-wo-distinct").innerHTML = "error";
//  }
};

getWODistinct.onerror = function() {
    document.getElementById("get-wo-distinct").innerHTML = "error";
};

getWODistinct.send();

// GET to different origin with CORS
var getDistinctCORS = new XMLHttpRequest();
getDistinctCORS.open('GET', 'http://localhost:9091/json?value=ok&cors=true', true);

getDistinctCORS.onload = function() {
  if (getDistinctCORS.status >= 200 && getDistinctCORS.status < 400) {
    var data = JSON.parse(getDistinctCORS.responseText);
    document.getElementById("get-distinct-cors").innerHTML = data.key;
  } else {
    document.getElementById("get-distinct-cors").innerHTML = "error";
  }
};

getDistinctCORS.onerror = function() {
    document.getElementById("get-distinct-cors").innerHTML = "error";
};

getDistinctCORS.send();


// GET to different origin with complex header
var getDistinctComplex = new XMLHttpRequest();
getDistinctComplex.open('GET', 'http://localhost:9091/json?value=ok', true);

getDistinctComplex.onload = function() {
  if (getDistinctComplex.status >= 200 && getDistinctComplex.status < 400) {
    var data = JSON.parse(getDistinctComplex.responseText);
    document.getElementById("get-distinct-complex").innerHTML = data.key;
  } else {
    document.getElementById("get-distinct-complex").innerHTML = "error";
  }
};

getDistinctComplex.onerror = function() {
    document.getElementById("get-distinct-complex").innerHTML = "error";
};

getDistinctComplex.setRequestHeader("complex","complex")
getDistinctComplex.send();


// GET to different origin with complex header
var getDistinctComplex = new XMLHttpRequest();
getDistinctComplex.open('GET', 'http://localhost:9091/json?value=ok&cors=true', true);

getDistinctComplex.onload = function() {
  if (getDistinctComplex.status >= 200 && getDistinctComplex.status < 400) {
    var data = JSON.parse(getDistinctComplex.responseText);
    document.getElementById("get-distinct-complex-cors").innerHTML = data.key;
  } else {
    document.getElementById("get-distinct-complex-cors").innerHTML = "error";
  }
};

getDistinctComplex.onerror = function() {
    document.getElementById("get-distinct-complex-cors").innerHTML = "error";
};

getDistinctComplex.setRequestHeader("complex","complex")
getDistinctComplex.send();

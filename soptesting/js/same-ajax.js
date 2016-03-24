// GET to same origin (with result)
var getSame = new XMLHttpRequest();
getSame.open('GET', 'http://localhost:9090/json?value=ok', true);

getSame.onload = function() {
  if (getSame.status >= 200 && getSame.status < 400) {
    var data = JSON.parse(getSame.responseText);
    document.getElementById("get-same").innerHTML = data.key;
  } else {
    document.getElementById("get-same").innerHTML = "error";
  }
};

getSame.onerror = function() {
    document.getElementById("get-same").innerHTML = "error";
};

getSame.send();
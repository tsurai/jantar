function get_elements_by_tag_and_attr(tag, attr) {
  var ret = new Array();
  var elements = document.getElementsByTagName(tag);
  
  for(i = 0; i < elements.length; i++) {
    if(elements[i].hasAttribute(attr)) {
      ret.push(elements[i]);
    }
  }

  return ret;
}

function get_csrf_token() {
  var elements = document.getElementsByName("csrf-token");

  for(i = 0; i < elements.length; i++) {
    if(elements[i].tagName.toLowerCase() == "meta" && elements[i].hasAttribute("content")) {
      return elements[i].getAttribute("content");
    }
  }

  return "";
}

function insert_csrf_token_into_forms(token) {
  if(token != "") {
    var elements = document.getElementsByTagName("form");

    for(i = 0; i < elements.length; i++) {
      elements[i].onsubmit = function(e) {
        this.innerHTML += '<input type="hidden" name="_csrf-token" value="'+token+'">';
      }
    }
  }
}

function insert_csrf_token_into_links(token) {
  if(token != "") {
    var elements = get_elements_by_tag_and_attr("a", "data-method");

    for(i = 0; i < elements.length; i++) {
      var href = elements[i].getAttribute("href");
      var method = elements[i].getAttribute("data-method").toUpperCase();
      
      if(!(method == "GET" || method == "POST" || method == "PUT")) {
        method = "POST";
      }

      elements[i].onclick = function(e) {
        var form = document.createElement("form");
        form.setAttribute("method", "POST");
        form.setAttribute("action", href);

        var input_method = document.createElement("input");
        input_method.setAttribute("type", "hidden");
        input_method.setAttribute("name", "_method");
        input_method.setAttribute("value", method);

        var input_csrf = document.createElement("input");
        input_csrf.setAttribute("type", "hidden");
        input_csrf.setAttribute("name", "_csrf-token");
        input_csrf.setAttribute("value", token);

        form.appendChild(input_method);
        form.appendChild(input_csrf);
        document.body.appendChild(form);

        e.preventDefault();
        form.submit();
      }
    }
  }  
}

function clickjacking_protection() {
  if (self === top) {
    var antiClickjack = document.getElementById("antiClickjack");
    antiClickjack.parentNode.removeChild(antiClickjack);
  } else {
    top.location = self.location;
  }
}

window.onload = function() {
  var csrf_token = get_csrf_token();
  if(csrf_token != "") {
    insert_csrf_token_into_links(csrf_token);
    insert_csrf_token_into_forms(csrf_token);
  }
}
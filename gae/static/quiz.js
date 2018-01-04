function getEntries(corpus) {
  const xhr = new XMLHttpRequest();
  const url = '/flashcards?n=10&corpus=' + corpus;
  xhr.responseType = 'json';
  xhr.onreadystatechange = function() {
    if (xhr.readyState === XMLHttpRequest.DONE) {
      runFlashcards(buildFlashcards(xhr.response));
    }
  };
  xhr.open('GET', url);
  xhr.send();
  document.getElementById("get-flashcards").hidden = true;
}

function runFlashcards(cards) {
  var div = document.createElement("div");
  document.getElementById("content").appendChild(div);
  div.style = "line-height: 30px; font-size: xx-large; text-align: center";
  var cur = 0;
  var space = vspace("50px");
  var status = document.createTextNode("")

  updateStatus = function() {
    status.textContent = (cur+1) + " / " + cards.length;
  };
  updateStatus();

  div.appendChild(vspace("100px"));
  div.appendChild(cards[cur]);
  div.appendChild(space);
  div.appendChild(button("Prev", function() {
    div.removeChild(cards[cur])
    cur--;
    if (cur < 0) {
      cur = cards.length-1;
    }
    div.insertBefore(cards[cur], space)
    updateStatus();
  }));
  div.appendChild(hspace());
  div.appendChild(button("Next", function() {
    div.removeChild(cards[cur])
    cur++;
    if (cur >= cards.length) {
      cur = 0
    }
    div.insertBefore(cards[cur], space)
    updateStatus();
  }));
  div.appendChild(vspace("10px"));
  div.appendChild(status);
}


function showEntries(entries) {
  var q = document.createElement("div");
    q.style="line-height: 30px";
  for (var i = 0; i < entries.length; i++) {
    q.appendChild(flashcard(entries[i]))
  }
  document.getElementById("questions").appendChild(q);
}

function buildFlashcards(entries) {
  var cards = []
  for (var i = 0; i < entries.length; i++) {
    cards.push(flashcard(entries[i]))
  }
  return cards
}


function flashcard(entry) {
  var card = document.createElement("div");
  //card.style = "text-align: center";
  var txt = document.createTextNode(entry.Question)
  var flip = button("Flip", function() {
    if (txt.textContent == entry.Question) {
      txt.textContent = entry.Answer;
    } else {
      txt.textContent = entry.Question;
    }
  })
  flip.style = "font-size: large";
  card.appendChild(txt);
  card.appendChild(vspace("30px"));
  card.appendChild(flip);
  return card
}


function button(text, func) {
    var b = document.createElement("button");
    b.innerHTML = text;
    b.onclick = func;
    return b;
}

function hspace() { return document.createTextNode(" "); }

function vspace(s) {
  var div = document.createElement("div");
  div.style = "height:" + s;
  return div
}

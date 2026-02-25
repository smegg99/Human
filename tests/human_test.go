// tests/human_test.go
package human_test

import (
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/smegg99/human"
)

const pageHTML = `<!DOCTYPE html>
<html><body>
  <div id="score">Clicks: 0</div>
  <div id="group" style="position:absolute">
    <input id="typebox" type="text" autocomplete="off" />
    <button id="target">Click me</button>
  </div>
  <div id="dot" style="position:fixed;width:8px;height:8px;border-radius:50%;background:red;pointer-events:none;z-index:9999;transform:translate(-50%,-50%)"></div>
  <script>
    let clicks = 0;
    const group = document.getElementById('group');
    const btn = document.getElementById('target');
    const typebox = document.getElementById('typebox');
    const score = document.getElementById('score');
    const dot = document.getElementById('dot');

    function moveGroup() {
      const x = Math.max(10, Math.floor(Math.random() * (window.innerWidth - 330)));
      const y = Math.max(10, Math.floor(Math.random() * (window.innerHeight - 60)));
      group.style.left = x + 'px';
      group.style.top = y + 'px';
    }

    group.style.left = (window.innerWidth / 2 - 160) + 'px';
    group.style.top = (window.innerHeight / 2 - 20) + 'px';

    btn.addEventListener('click', () => {
      clicks++;
      score.textContent = 'Clicks: ' + clicks;
      if (typebox.value) { typebox.value = ''; }
      moveGroup();
    });

    document.addEventListener('mousemove', (e) => {
      dot.style.left = e.clientX + 'px';
      dot.style.top = e.clientY + 'px';
    });
  </script>
</body></html>`

var loremWords = strings.Fields("lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua")

func randomPhrase(n int) string {
	words := make([]string, n)
	for i := range words {
		words[i] = loremWords[rand.Intn(len(loremWords))]
	}
	return strings.Join(words, " ")
}

func TestHumanCursor(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(pageHTML))
	}))
	defer srv.Close()

	u := launcher.New().Headless(false).MustLaunch()
	browser := rod.New().ControlURL(u).MustConnect()
	defer browser.MustClose()

	presets := []struct {
		name string
		opt  human.Option
	}{
		{"Fast", human.Fast()},
		{"Swift", human.Swift()},
		{"Casual", human.Casual()},
		{"Beginner", human.Beginner()},
	}

	for _, p := range presets {
		t.Run(p.name, func(t *testing.T) {
			page := browser.MustPage(srv.URL).MustWaitStable()
			defer page.Close()
			cursor := human.New(page, p.opt)

			for i := 1; i <= 3; i++ {
				typebox := page.MustElement("#typebox")
				if err := cursor.Click(typebox); err != nil {
					t.Fatalf("round %d: click input: %v", i, err)
				}

				phrase := randomPhrase(3 + rand.Intn(3))
				cursor.Type(phrase)
				time.Sleep(50 * time.Millisecond)

				got := page.MustEval(`() => document.getElementById('typebox').value`).String()
				if got != phrase {
					t.Errorf("round %d: typed %q, got %q", i, phrase, got)
				}

				btn := page.MustElement("#target")
				if err := cursor.Click(btn); err != nil {
					t.Fatalf("round %d: click button: %v", i, err)
				}
				time.Sleep(50 * time.Millisecond)

				score := page.MustEval(`() => document.getElementById('score').textContent`).String()
				if expected := fmt.Sprintf("Clicks: %d", i); score != expected {
					t.Errorf("round %d: expected %q, got %q", i, expected, score)
				}
			}

			typebox := page.MustElement("#typebox")
			cursor.Click(typebox)
			cursor.Type("test")
			time.Sleep(50 * time.Millisecond)
			cursor.KeyCombo([]input.Key{input.ControlLeft}, input.KeyA)
			time.Sleep(50 * time.Millisecond)
			cursor.PressKey(input.Backspace)
			time.Sleep(50 * time.Millisecond)
			val := page.MustEval(`() => document.getElementById('typebox').value`).String()
			if val != "" {
				t.Errorf("Ctrl+A + Backspace: expected empty, got %q", val)
			}
		})
	}
}

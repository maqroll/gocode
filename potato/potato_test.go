package potato

import (
	"flag"
	gm "github.com/onsi/gomega"
	"github.com/sclevine/agouti"
	am "github.com/sclevine/agouti/matchers"
	"os"
	"testing"
)

var terminal = flag.String("terminal", "", "terminal")
var shouldRunAsManager = flag.Bool("manager", false, "should run as manager?")
var app = flag.String("app", "", "app")

func TestMain(m *testing.M) {
	flag.Parse()
	os.Exit(m.Run())
}

func TestApp(t *testing.T) {
	gm.RegisterTestingT(t)
	// driver = agouti.PhantomJS()
	// driver = agouti.Selenium()

	gm.Expect(*terminal).ShouldNot(gm.Equal(""), "Terminal should not be empty")
	gm.Expect(*app).ShouldNot(gm.Equal(""), "App should not be empty")

	// chromedriver.exe should be in %PATH%
	driver := agouti.ChromeDriver(agouti.Desired(agouti.Capabilities{
		"chromeOptions": map[string]string{
			"binary": "D:\\bin\\PortableApps\\GoogleChromePortable\\GoogleChromePortable.exe",
		},
	}))

	gm.Expect(driver.Start()).To(gm.Succeed(), "Driver didn't start")

	page, err := driver.NewPage()
	gm.Expect(err).NotTo(gm.HaveOccurred(), "Failed to create a new page")

	gm.Expect(page.Navigate("http://www.google.com")).To(gm.Succeed(), "Failed to navigate to google")

	gm.Expect(page.URL()).To(gm.ContainSubstring("https://www.google.es"), "Failed to check url")

	input := page.Find("#lst-ib")
	gm.Expect(input).To(am.BeFound(), "Failed to found input")

	gm.Expect(input.Fill("prueba")).To(gm.Succeed(), "Failed to fill input")

	form := page.Find("#tsf")
	gm.Expect(form).To(am.BeFound(), "Failed to find form")

	gm.Expect(form.Submit()).To(gm.Succeed(), "Failed to submit form")

	gm.Eventually(func() *agouti.Selection {
		return page.Find("#appbar")
	}, "1m", "1s").Should(am.BeFound())

	results := page.All(".g")

	gm.Expect(results).To(am.BeFound(), "Failed to find results")

	// Kill the process! If portable, not effective at all
	gm.Expect(driver.Stop()).To(gm.Succeed(), "Failed to stop driver")
}

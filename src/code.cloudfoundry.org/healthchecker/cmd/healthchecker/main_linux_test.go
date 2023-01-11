package main_test

import (
	"net"
	"os"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("HealthChecker - Linux Specific", func() {
	BeforeEach(HealthCheckerBeforeEach)
	JustBeforeEach(HealthCheckerJustBeforeEach)
	AfterEach(HealthCheckerAfterEach)

	Context("when there is a unix socket server running", func() {
		var server *ghttp.Server
		BeforeEach(func() {
			unixSocket, err := os.CreateTemp("", "ghttpUnixSocket.*")
			Expect(err).NotTo(HaveOccurred())
			err = os.Remove(unixSocket.Name())
			Expect(err).NotTo(HaveOccurred())

			unixListener, err := net.Listen("unix", unixSocket.Name())
			Expect(err).NotTo(HaveOccurred())

			server = ghttp.NewUnstartedServer()
			server.HTTPTestServer.Listener = unixListener
			server.RouteToHandler(
				"GET", "/some-path",
				ghttp.RespondWith(200, "ok"),
			)
			server.Start()

			cfg.HealthCheckEndpoint.Socket = unixSocket.Name()
			cfg.LogLevel = "debug"
			cfg.HealthCheckEndpoint.Path = "/some-path"
			cfg.StartupDelayBuffer = 5 * time.Second
			cfg.HealthCheckPollInterval = 500 * time.Millisecond
			cfg.HealthCheckTimeout = 5 * time.Second
		})

		AfterEach(func() {
			server.Close()
		})

		It("works", func() {
			Eventually(session.Out, 10*time.Second).Should(gbytes.Say("Verifying endpoint"))
			Eventually(func() int { return len(server.ReceivedRequests()) }, 10*time.Second).Should(BeNumerically(">", 0))
		})
	})
})

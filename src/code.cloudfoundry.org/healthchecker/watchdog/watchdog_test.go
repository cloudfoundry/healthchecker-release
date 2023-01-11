package watchdog_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"time"

	"code.cloudfoundry.org/cf-networking-helpers/healthchecker/watchdog"
	"code.cloudfoundry.org/lager"
	"code.cloudfoundry.org/lager/lagertest"
	"github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo"
	"github.com/onsi/gomega"
	. "github.com/onsi/gomega"
)

var _ = Describe("Watchdog", func() {
	var (
		srv                *http.Server
		dog                *watchdog.Watchdog
		addr               string
		pollInterval       time.Duration
		healthcheckTimeout time.Duration
		logger             lager.Logger
		failureCounterFile *os.File
		err                error
		u                  *url.URL
	)

	healthcheckTimeout = 500 * time.Millisecond
	runServer := func(httpHandler http.Handler) *http.Server {
		localSrv := http.Server{
			Addr:    addr,
			Handler: httpHandler,
		}
		go func() {
			defer GinkgoRecover()
			err := localSrv.ListenAndServe()
			Expect(err).To(HaveOccurred())
			Expect(err).To(MatchError(http.ErrServerClosed))
		}()
		Eventually(func() error {
			_, err := http.Get(fmt.Sprintf("http://%s", addr))
			return err
		}).Should(Not(HaveOccurred()))
		return &localSrv
	}

	BeforeEach(func() {
		addr = fmt.Sprintf("localhost:%d", 9850+ginkgo.GinkgoParallelProcess())
		pollInterval = 10 * time.Millisecond
		logger = lagertest.NewTestLogger("watchdog")
	})

	Context("Starting WatchDog from scratch", func() {
		var failureCounterFileName string

		JustBeforeEach(func() {
			failureCounterFileName = "potatoes"
			u, err = url.Parse("http://" + addr + "/healthz")
			Expect(err).ToNot(HaveOccurred())
		})

		AfterEach(func() {
			err = os.Remove(failureCounterFileName)
			Expect(err).ToNot(HaveOccurred())
		})

		It("creates the FailureCounter file", func() {
			_, err := os.Stat(failureCounterFileName)
			Expect(err).To(HaveOccurred())

			dog = watchdog.NewWatchdog(u, "some-component", failureCounterFileName, pollInterval, healthcheckTimeout, logger)
			Expect(dog).ToNot(BeNil())

			_, err = os.ReadFile(failureCounterFileName)
			Expect(err).ToNot(HaveOccurred())
		})
	})

	Context("Operating", func() {
		BeforeEach(func() {
			var err error
			failureCounterFile, err = os.CreateTemp("", "ginkgoWatchdogFailureCountFile.*")
			Expect(err).NotTo(HaveOccurred())
		})

		JustBeforeEach(func() {
			u, err := url.Parse("http://" + addr + "/healthz")
			Expect(err).NotTo(HaveOccurred())
			dog = watchdog.NewWatchdog(u, "some-component", failureCounterFile.Name(), pollInterval, healthcheckTimeout, logger)
		})

		AfterEach(func() {
			ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
			defer cancel()
			srv.Shutdown(ctx)
			srv.Close()
		})

		Context("HitHealthcheckEndpoint", func() {
			var statusCode int
			BeforeEach(func() {
				httpHandler := http.NewServeMux()
				httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
					rw.WriteHeader(statusCode)
					r.Close = true
				})
				srv = runServer(httpHandler)
			})
			It("does not return an error if the endpoint returns a 200", func() {
				statusCode = http.StatusOK
				err := dog.HitHealthcheckEndpoint()
				Expect(err).NotTo(HaveOccurred())
			})

			Context("when endpoint does not respond with a 200", func() {
				It("returns an error if the endpoint", func() {
					statusCode = http.StatusServiceUnavailable
					err := dog.HitHealthcheckEndpoint()
					Expect(err).To(HaveOccurred())
				})
			})
		})

		Context("WatchHealthcheckEndpoint", func() {
			var signals chan os.Signal

			BeforeEach(func() {
				signals = make(chan os.Signal)
				err := os.WriteFile(failureCounterFile.Name(), []byte("8\n"), 0644)
				Expect(err).NotTo(HaveOccurred())
			})

			Context("the healthcheck passes repeatedly", func() {
				BeforeEach(func() {
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						rw.WriteHeader(http.StatusOK)
						r.Close = true
					})
					srv = runServer(httpHandler)
				})

				It("does not return an error", func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*pollInterval)
					defer cancel()
					err := dog.WatchHealthcheckEndpoint(ctx, signals)
					Expect(err).NotTo(HaveOccurred())
				})

				It("resets the value in failureCounterFile", func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*pollInterval)
					defer cancel()
					dog.WatchHealthcheckEndpoint(ctx, signals)
					Eventually(func(g gomega.Gomega) {
						content, err := os.ReadFile(failureCounterFile.Name())
						g.Expect(err).NotTo(HaveOccurred())
						g.Expect(string(content)).To(Equal("0\n"))
					}).Should(Succeed())
				})
			})

			Context("the healthcheck first passes, and subsequently fails", func() {
				BeforeEach(func() {
					var visitCount int
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						if visitCount == 0 {
							rw.WriteHeader(http.StatusOK)
						} else {
							rw.WriteHeader(http.StatusNotAcceptable)
						}
						r.Close = true
						visitCount++
					})
					srv = runServer(httpHandler)
				})

				It("returns an error", func() {
					err := dog.WatchHealthcheckEndpoint(context.Background(), signals)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("the healthcheck fails repeatedly", func() {
				var retriesNum int

				BeforeEach(func() {
					retriesNum = 0
					httpHandler := http.NewServeMux()

					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						rw.WriteHeader(http.StatusNotAcceptable)
						retriesNum++
						r.Close = true
					})
					srv = runServer(httpHandler)
				})

				It("retries 3 times and then fails", func() {
					err := dog.WatchHealthcheckEndpoint(context.Background(), signals)
					Expect(err).To(HaveOccurred())
					Expect(retriesNum).To(Equal(3))
				})

				It("increments the failureCounterFile", func() {
					dog.WatchHealthcheckEndpoint(context.Background(), signals)
					content, err := os.ReadFile(failureCounterFile.Name())
					Expect(err).NotTo(HaveOccurred())
					Expect(string(content)).To(Equal("9\n"))
				})
			})

			Context("the healthcheck fails and then succeeds", func() {
				BeforeEach(func() {
					var visitCount int
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						if visitCount < 2 {
							rw.WriteHeader(http.StatusNotAcceptable)
						} else {
							rw.WriteHeader(http.StatusOK)
						}
						r.Close = true
						visitCount++
					})
					srv = runServer(httpHandler)
				})

				It("retries on failures and then succeeds", func() {
					ctx, cancel := context.WithTimeout(context.Background(), 10*pollInterval)
					defer cancel()
					err := dog.WatchHealthcheckEndpoint(ctx, signals)
					Expect(err).NotTo(HaveOccurred())
				})
			})

			Context("the healthcheck fails and then succeeds, and then fails again", func() {
				var firstRetriesNum int
				var secondRetriesNum int

				BeforeEach(func() {
					var visitCount int
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						if visitCount < 2 {
							rw.WriteHeader(http.StatusNotAcceptable)
							firstRetriesNum++
						} else if visitCount < 4 {
							rw.WriteHeader(http.StatusOK)
						} else {
							rw.WriteHeader(http.StatusNotAcceptable)
							secondRetriesNum++
						}
						r.Close = true
						visitCount++
					})
					srv = runServer(httpHandler)
				})

				It("retries on second failures", func() {
					err := dog.WatchHealthcheckEndpoint(context.Background(), signals)
					Expect(err).To(HaveOccurred())
					Expect(firstRetriesNum).To(Equal(2))
					Expect(secondRetriesNum).To(Equal(3))
				})
			})

			Context("the endpoint does not respond in the configured timeout", func() {
				BeforeEach(func() {
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						time.Sleep(5 * healthcheckTimeout)
						rw.WriteHeader(http.StatusOK)
						r.Close = true
					})
					srv = runServer(httpHandler)
				})

				It("returns an error", func() {
					ctx, cancel := context.WithTimeout(context.Background(), 100*healthcheckTimeout)
					defer cancel()
					err := dog.WatchHealthcheckEndpoint(ctx, signals)
					Expect(err).To(HaveOccurred())
				})
			})

			Context("context is canceled", func() {
				var ctx context.Context
				var visitCount int

				BeforeEach(func() {
					var cancel context.CancelFunc
					ctx, cancel = context.WithCancel(context.Background())
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						rw.WriteHeader(http.StatusOK)
						r.Close = true
						visitCount++
						if visitCount == 3 {
							cancel()
						}
					})
					srv = runServer(httpHandler)
				})

				It("stops the healthchecker", func() {
					err := dog.WatchHealthcheckEndpoint(ctx, signals)
					Expect(err).NotTo(HaveOccurred())
					Expect(visitCount).To(Equal(3))
				})
			})

			Context("received the TEST_SIGNAL signal", func() {
				var visitCount int

				BeforeEach(func() {
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						rw.WriteHeader(http.StatusOK)
						r.Close = true
						visitCount++
						if visitCount == 3 {
							go func() {
								signals <- TEST_SIGNAL
							}()
						}
					})
					srv = runServer(httpHandler)
				})

				It("stops the healthchecker without an error", func() {
					err := dog.WatchHealthcheckEndpoint(context.Background(), signals)
					Expect(err).NotTo(HaveOccurred())
					Expect(visitCount).To(Equal(3))
				})
			})

			Context("gorouter exited before we received a TEST_SIGNAL signal", func() {
				BeforeEach(func() {
					httpHandler := http.NewServeMux()
					httpHandler.HandleFunc("/healthz", func(rw http.ResponseWriter, r *http.Request) {
						rw.WriteHeader(http.StatusServiceUnavailable)
						r.Close = true
						go func() {
							signals <- TEST_SIGNAL
						}()
					})
					srv = runServer(httpHandler)
				})

				It("stops the healthchecker without an error", func() {
					err := dog.WatchHealthcheckEndpoint(context.Background(), signals)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})
})

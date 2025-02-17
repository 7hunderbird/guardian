package runrunc_test

import (
	"errors"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/opencontainers/runtime-spec/specs-go"

	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/guardian/rundmc/depot"
	"code.cloudfoundry.org/guardian/rundmc/goci"
	. "code.cloudfoundry.org/guardian/rundmc/runrunc"
	"code.cloudfoundry.org/guardian/rundmc/runrunc/runruncfakes"
	"code.cloudfoundry.org/lager/lagertest"
)

var _ = Describe("Infoer", func() {
	var (
		infoer    *Infoer
		fakeDepot *runruncfakes.FakeDepot
		logger    *lagertest.TestLogger
	)

	BeforeEach(func() {
		logger = lagertest.NewTestLogger("test")
		fakeDepot = new(runruncfakes.FakeDepot)
		infoer = NewInfoer(fakeDepot)
	})

	Describe("BundleInfo", func() {
		var (
			bundlePath string
			bundle     goci.Bndl
			err        error
		)

		BeforeEach(func() {
			fakeDepot.LookupReturns("/the/bundle/path", nil)
			fakeDepot.LoadReturns(goci.Bndl{Spec: specs.Spec{Version: "my-bundle"}}, nil)
		})

		JustBeforeEach(func() {
			bundlePath, bundle, err = infoer.BundleInfo(logger, "my-container")
		})

		It("returns the bundle for the specified container", func() {
			Expect(err).NotTo(HaveOccurred())
			Expect(bundlePath).To(Equal("/the/bundle/path"))
			Expect(bundle.Spec.Version).To(Equal("my-bundle"))

			Expect(fakeDepot.LookupCallCount()).To(Equal(1))
			lookupLogger, lookupHandle := fakeDepot.LookupArgsForCall(0)
			Expect(lookupLogger).To(Equal(logger))
			Expect(lookupHandle).To(Equal("my-container"))

			Expect(fakeDepot.LoadCallCount()).To(Equal(1))
			loadLogger, loadHandle := fakeDepot.LoadArgsForCall(0)
			Expect(loadLogger).To(Equal(logger))
			Expect(loadHandle).To(Equal("my-container"))
		})

		When("the container does not exist", func() {
			BeforeEach(func() {
				fakeDepot.LookupReturns("", depot.ErrDoesNotExist)
			})

			It("returns a garden.ContainerNotFoundError", func() {
				Expect(err).To(Equal(garden.ContainerNotFoundError{Handle: "my-container"}))
			})
		})

		When("looking up the bundle path fails", func() {
			BeforeEach(func() {
				fakeDepot.LookupReturns("", errors.New("lookup-error"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("lookup-error"))
			})
		})

		When("loading the bundle path fails", func() {
			BeforeEach(func() {
				fakeDepot.LoadReturns(goci.Bndl{}, errors.New("load-error"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("load-error"))
			})
		})
	})
})

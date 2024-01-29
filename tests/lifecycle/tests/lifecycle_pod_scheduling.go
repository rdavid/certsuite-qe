package tests

import (
	"github.com/golang/glog"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/test-network-function/cnfcert-tests-verification/tests/globalhelper"
	"github.com/test-network-function/cnfcert-tests-verification/tests/globalparameters"
	"github.com/test-network-function/cnfcert-tests-verification/tests/utils/config"
	"github.com/test-network-function/cnfcert-tests-verification/tests/utils/daemonset"
	"github.com/test-network-function/cnfcert-tests-verification/tests/utils/deployment"

	tshelper "github.com/test-network-function/cnfcert-tests-verification/tests/lifecycle/helper"
	tsparams "github.com/test-network-function/cnfcert-tests-verification/tests/lifecycle/parameters"
)

var _ = Describe("lifecycle-pod-scheduling", func() {
	var randomNamespace string
	var randomReportDir string
	var randomTnfConfigDir string

	configSuite, err := config.NewConfig()
	if err != nil {
		glog.Fatalf("can not load config file: %w", err)
	}

	BeforeEach(func() {
		// Create random namespace and keep original report and TNF config directories
		randomNamespace, randomReportDir, randomTnfConfigDir = globalhelper.BeforeEachSetupWithRandomNamespace(tsparams.LifecycleNamespace)

		By("Define TNF config file")
		err = globalhelper.DefineTnfConfig(
			[]string{randomNamespace},
			[]string{tsparams.TestPodLabel},
			[]string{tsparams.TnfTargetOperatorLabels},
			[]string{},
			[]string{}, randomTnfConfigDir)
		Expect(err).ToNot(HaveOccurred())
	})

	AfterEach(func() {
		globalhelper.AfterEachCleanupWithRandomNamespace(randomNamespace, randomReportDir, randomTnfConfigDir, tsparams.WaitingTime)
	})

	// 48120
	It("One deployment, no nodeSelector nor nodeAffinity", func() {
		By("Define Deployment")
		deployment, err := tshelper.DefineDeployment(1, 1, tsparams.TestDeploymentName, randomNamespace)
		Expect(err).ToNot(HaveOccurred())

		By("Create Deployment")
		err = globalhelper.CreateAndWaitUntilDeploymentIsReady(deployment, tsparams.WaitingTime)
		Expect(err).ToNot(HaveOccurred())

		By("Start lifecycle-pod-scheduling test")
		err = globalhelper.LaunchTests(tsparams.TnfPodSchedulingTcName,
			globalhelper.ConvertSpecNameToFileName(CurrentSpecReport().FullText()), randomReportDir, randomTnfConfigDir)
		Expect(err).ToNot(HaveOccurred())

		if globalhelper.GetNumberOfNodes(globalhelper.GetAPIClient().K8sClient.CoreV1()) == 1 {
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCaseSkipped, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		} else {
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCasePassed, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	// 48458
	It("One deployment with nodeSelector [negative]", func() {
		By("Define Deployment with nodeSelector")
		deploymenta, err := tshelper.DefineDeployment(1, 1, tsparams.TestDeploymentName, randomNamespace)
		Expect(err).ToNot(HaveOccurred())

		deployment.RedefineWithNodeSelector(deploymenta, map[string]string{configSuite.General.CnfNodeLabel: ""})
		Expect(err).ToNot(HaveOccurred())

		By("Create Deployment")
		err = globalhelper.CreateAndWaitUntilDeploymentIsReady(deploymenta, tsparams.WaitingTime)
		Expect(err).ToNot(HaveOccurred())

		By("Start lifecycle-pod-scheduling test")
		err = globalhelper.LaunchTests(tsparams.TnfPodSchedulingTcName,
			globalhelper.ConvertSpecNameToFileName(CurrentSpecReport().FullText()), randomReportDir, randomTnfConfigDir)

		if globalhelper.GetNumberOfNodes(globalhelper.GetAPIClient().K8sClient.CoreV1()) == 1 {
			Expect(err).ToNot(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCaseSkipped, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCaseFailed, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	// 48470
	It("One deployment with nodeAffinity [negative]", func() {
		By("Define Deployment with nodeAffinity")
		deploymenta, err := tshelper.DefineDeployment(1, 1, tsparams.TestDeploymentName, randomNamespace)
		Expect(err).ToNot(HaveOccurred())

		deployment.RedefineWithNodeAffinity(deploymenta, configSuite.General.CnfNodeLabel)

		By("Create Deployment")
		err = globalhelper.CreateAndWaitUntilDeploymentIsReady(deploymenta, tsparams.WaitingTime)
		Expect(err).ToNot(HaveOccurred())

		By("Start lifecycle-pod-scheduling test")
		err = globalhelper.LaunchTests(tsparams.TnfPodSchedulingTcName,
			globalhelper.ConvertSpecNameToFileName(CurrentSpecReport().FullText()), randomReportDir, randomTnfConfigDir)

		if globalhelper.GetNumberOfNodes(globalhelper.GetAPIClient().K8sClient.CoreV1()) == 1 {
			Expect(err).ToNot(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCaseSkipped, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCaseFailed, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	// 48471
	It("Two deployments, one pod each, one pod with nodeAffinity [negative]", func() {

		By("Define Deployment without nodeAffinity")
		deploymenta, err := tshelper.DefineDeployment(1, 1, tsparams.TestDeploymentName, randomNamespace)
		Expect(err).ToNot(HaveOccurred())

		By("Create Deployment")
		err = globalhelper.CreateAndWaitUntilDeploymentIsReady(deploymenta, tsparams.WaitingTime)
		Expect(err).ToNot(HaveOccurred())

		By("Define Deployment with nodeAffinity")
		deploymentb, err := tshelper.DefineDeployment(1, 1, "lifecycle-dpb", randomNamespace)
		Expect(err).ToNot(HaveOccurred())

		deployment.RedefineWithNodeAffinity(deploymentb, configSuite.General.CnfNodeLabel)

		By("Create Deployment")
		err = globalhelper.CreateAndWaitUntilDeploymentIsReady(deploymentb, tsparams.WaitingTime)
		Expect(err).ToNot(HaveOccurred())

		By("Start lifecycle-pod-scheduling test")
		err = globalhelper.LaunchTests(tsparams.TnfPodSchedulingTcName,
			globalhelper.ConvertSpecNameToFileName(CurrentSpecReport().FullText()), randomReportDir, randomTnfConfigDir)

		if globalhelper.GetNumberOfNodes(globalhelper.GetAPIClient().K8sClient.CoreV1()) == 1 {
			Expect(err).ToNot(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCaseSkipped, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(tsparams.TnfPodSchedulingTcName, globalparameters.TestCaseFailed, randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	// 48472
	It("One deployment, one daemonSet [negative]", func() {

		By("Define Deployment without nodeAffinity/ nodeSelector")
		deployment, err := tshelper.DefineDeployment(1, 1, tsparams.TestDeploymentName, randomNamespace)
		Expect(err).ToNot(HaveOccurred())

		By("Create Deployment")
		err = globalhelper.CreateAndWaitUntilDeploymentIsReady(deployment, tsparams.WaitingTime)
		Expect(err).ToNot(HaveOccurred())

		By("Define daemonSet")
		daemonSet := daemonset.DefineDaemonSet(randomNamespace,
			globalhelper.GetConfiguration().General.TestImage,
			tsparams.TestTargetLabels, tsparams.TestDaemonSetName)

		By("Create daemonSet")
		err = globalhelper.CreateAndWaitUntilDaemonSetIsReady(daemonSet, tsparams.WaitingTime)
		Expect(err).ToNot(HaveOccurred())

		By("Start lifecycle-pod-scheduling test")
		err = globalhelper.LaunchTests(tsparams.TnfPodSchedulingTcName,
			globalhelper.ConvertSpecNameToFileName(CurrentSpecReport().FullText()), randomReportDir, randomTnfConfigDir)

		if globalhelper.GetNumberOfNodes(globalhelper.GetAPIClient().K8sClient.CoreV1()) == 1 {
			Expect(err).ToNot(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(
				tsparams.TnfPodSchedulingTcName,
				globalparameters.TestCaseSkipped,
				randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		} else {
			Expect(err).To(HaveOccurred())
			By("Verify test case status in Claim report")
			err = globalhelper.ValidateIfReportsAreValid(
				tsparams.TnfPodSchedulingTcName,
				globalparameters.TestCaseFailed,
				randomReportDir)
			Expect(err).ToNot(HaveOccurred())
		}
	})
})

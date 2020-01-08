package ssmparameter

import (
	"context"

	ssmparameterv1alpha1 "github.com/calebpalmer/aws-ssm-secret-operator/pkg/apis/ssmparameter/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"

	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"os"
)

var log = logf.Log.WithName("controller_ssmparameter")

// Add creates a new SSMParameter Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSSMParameter{
		client:     mgr.GetClient(),
		scheme:     mgr.GetScheme(),
		ssmService: makeSSMSession(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ssmparameter-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource SSMParameter
	err = c.Watch(&source.Kind{Type: &ssmparameterv1alpha1.SSMParameter{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner SSMParameter
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &ssmparameterv1alpha1.SSMParameter{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSSMParameter implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSSMParameter{}

// ReconcileSSMParameter reconciles a SSMParameter object
type ReconcileSSMParameter struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client     client.Client
	scheme     *runtime.Scheme
	ssmService *ssm.SSM
}

// Reconcile reads that state of the cluster for a SSMParameter object and makes changes based on the state read
// and what is in the SSMParameter.Spec
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSSMParameter) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the SSMParameter instance
	instance := &ssmparameterv1alpha1.SSMParameter{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// log some stuff
	msg := fmt.Sprintf("Reconciling SSMParameter with path %s", instance.Spec.Path)
	reqLogger.Info(msg)

	// get the secret value
	if err != nil {
		reqLogger.Error(err, "Error reading SSM Parameter Value")
		return reconcile.Result{}, err
	}

	// Check if this Secret already exists
	found := &corev1.Secret{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.Name, Namespace: request.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		// create the secret
		secret, err := r.secretForSSMParameter(instance, request.Namespace)
		if err != nil {
			return reconcile.Result{}, err
		}

		reqLogger.Info("Creating a new Secret", "Deployment.Namespace", "", "Deployment.Name", "")

		err = r.client.Create(context.TODO(), secret)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Secret", "Secret.Namespace", secret.Namespace, "Secret.Name", secret.Name)
			return reconcile.Result{}, err
		}

		// Secret created successfully - return and requeue
		return reconcile.Result{}, nil
	}

	return reconcile.Result{}, nil
}

// secretForSSMParameter returns a Secret object for an SSM Parameter value.
func (r *ReconcileSSMParameter) secretForSSMParameter(param *ssmparameterv1alpha1.SSMParameter, namespace string) (*corev1.Secret, error) {
	// get the secret value
	value, err := r.readSSMParameter(param.Spec.Path, param.Spec.Decrypt)
	if err != nil {
		return nil, err
	}

	// make the secret
	stringData := make(map[string]string)
	stringData["value"] = value

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      param.Spec.Name,
			Namespace: namespace,
		},
		StringData: stringData,
	}

	// Set SSMParameter instance as the owner and controller
	controllerutil.SetControllerReference(param, secret, r.scheme)
	return secret, nil
}

// makeSSMSession creates and returns an AWS SSM session.
func makeSSMSession() *ssm.SSM {
	// create the session
	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}

	sess, err := session.NewSessionWithOptions(session.Options{
		Config:            aws.Config{Region: aws.String(region)},
		SharedConfigState: session.SharedConfigEnable,
	})

	if err != nil {
		panic(err)
	}

	// get the ssm service
	return ssm.New(sess)
}

// readSSMParameter reads and returns the value for a given AWS Systems Manager Parameter Store.
func (r *ReconcileSSMParameter) readSSMParameter(name string, decrypt bool) (string, error) {
	parm, err := r.ssmService.GetParameter(&ssm.GetParameterInput{
		Name:           &name,
		WithDecryption: &decrypt,
	})

	if err != nil {
		return "", err
	}

	return *parm.Parameter.Value, nil
}

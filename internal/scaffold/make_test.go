package scaffold

import (
	"os"
	"strings"
	"testing"
)

func TestMakeControllerWithPathPrefix(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeController("Api/UserController")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Http/Controllers/Api/UserController.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("controller with path prefix not created")
	}

	content, _ := os.ReadFile(path)
	if !strings.Contains(string(content), "type UserController struct") {
		t.Error("controller does not contain correct struct name")
	}
}

func TestMakeControllerAPI(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeController("UserController", true)
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Http/Controllers/UserController.go"
	content, _ := os.ReadFile(path)
	contentStr := string(content)

	// Should have Index, Store, Show, Update, Destroy
	if !strings.Contains(contentStr, "func (c *UserController) Index") {
		t.Error("API controller missing Index method")
	}
	if !strings.Contains(contentStr, "func (c *UserController) Store") {
		t.Error("API controller missing Store method")
	}
	if !strings.Contains(contentStr, "func (c *UserController) Show") {
		t.Error("API controller missing Show method")
	}
	if !strings.Contains(contentStr, "func (c *UserController) Update") {
		t.Error("API controller missing Update method")
	}
	if !strings.Contains(contentStr, "func (c *UserController) Destroy") {
		t.Error("API controller missing Destroy method")
	}

	// Should NOT have Create or Edit
	if strings.Contains(contentStr, "func (c *UserController) Create") {
		t.Error("API controller should not have Create method")
	}
	if strings.Contains(contentStr, "func (c *UserController) Edit") {
		t.Error("API controller should not have Edit method")
	}
}

func TestMakeControllerResourceful(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeController("PostController", false)
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Http/Controllers/PostController.go"
	content, _ := os.ReadFile(path)
	contentStr := string(content)

	// Should have all resourceful methods
	expectedMethods := []string{"Index", "Create", "Store", "Show", "Edit", "Update", "Destroy"}
	for _, method := range expectedMethods {
		if !strings.Contains(contentStr, "func (c *PostController) "+method) {
			t.Errorf("Resourceful controller missing %s method", method)
		}
	}
}

func TestMakeRequest(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeRequest("CreateUser")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Http/Requests/CreateUser.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("request not created")
	}

	content, _ := os.ReadFile(path)
	contentStr := string(content)

	// Check for required methods
	if !strings.Contains(contentStr, "type CreateUserRequest struct") {
		t.Error("request does not contain correct struct")
	}
	if !strings.Contains(contentStr, "func (r *CreateUserRequest) Rules()") {
		t.Error("request missing Rules method")
	}
	if !strings.Contains(contentStr, "func (r *CreateUserRequest) Authorize()") {
		t.Error("request missing Authorize method")
	}
	if !strings.Contains(contentStr, "func (r *CreateUserRequest) Messages()") {
		t.Error("request missing Messages method")
	}
}

func TestMakeRequestWithPathPrefix(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeRequest("Api/UpdateProfile")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Http/Requests/Api/UpdateProfile.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("request with path prefix not created")
	}
}

func TestMakeCommand(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeCommand("SendEmails")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Console/Commands/SendEmails.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("command not created")
	}

	content, _ := os.ReadFile(path)
	contentStr := string(content)

	// Check for required methods
	if !strings.Contains(contentStr, "type SendEmails struct") {
		t.Error("command does not contain correct struct")
	}
	if !strings.Contains(contentStr, "func (c *SendEmails) Signature()") {
		t.Error("command missing Signature method")
	}
	if !strings.Contains(contentStr, "func (c *SendEmails) Description()") {
		t.Error("command missing Description method")
	}
	if !strings.Contains(contentStr, "func (c *SendEmails) Handle()") {
		t.Error("command missing Handle method")
	}
}

func TestMakeEvent(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeEvent("UserRegistered")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Events/UserRegistered.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("event not created")
	}

	content, _ := os.ReadFile(path)
	contentStr := string(content)

	if !strings.Contains(contentStr, "type UserRegistered struct") {
		t.Error("event does not contain correct struct")
	}
	if !strings.Contains(contentStr, "package Events") {
		t.Error("event has incorrect package")
	}
}

func TestMakeEventWithPathPrefix(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeEvent("Order/OrderPlaced")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Events/Order/OrderPlaced.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("event with path prefix not created")
	}
}

func TestMakeListener(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeListener("SendWelcomeEmail")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Listeners/SendWelcomeEmail.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("listener not created")
	}

	content, _ := os.ReadFile(path)
	contentStr := string(content)

	// Check for required methods
	if !strings.Contains(contentStr, "type SendWelcomeEmail struct") {
		t.Error("listener does not contain correct struct")
	}
	if !strings.Contains(contentStr, "func (l *SendWelcomeEmail) Handle") {
		t.Error("listener missing Handle method")
	}
	if !strings.Contains(contentStr, "func (l *SendWelcomeEmail) ShouldQueue") {
		t.Error("listener missing ShouldQueue method")
	}
}

func TestMakeJob(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeJob("ProcessVideo")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Jobs/ProcessVideo.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("job not created")
	}

	content, _ := os.ReadFile(path)
	contentStr := string(content)

	// Check for required methods
	if !strings.Contains(contentStr, "type ProcessVideo struct") {
		t.Error("job does not contain correct struct")
	}
	if !strings.Contains(contentStr, "func (j *ProcessVideo) Handle()") {
		t.Error("job missing Handle method")
	}
	if !strings.Contains(contentStr, "func (j *ProcessVideo) Queue()") {
		t.Error("job missing Queue method")
	}
	if !strings.Contains(contentStr, "func (j *ProcessVideo) Tries()") {
		t.Error("job missing Tries method")
	}
	if !strings.Contains(contentStr, "func (j *ProcessVideo) Timeout()") {
		t.Error("job missing Timeout method")
	}
	if !strings.Contains(contentStr, "import \"time\"") {
		t.Error("job missing time import")
	}
}

func TestMakeJobWithPathPrefix(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakeJob("Media/ConvertVideo")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Jobs/Media/ConvertVideo.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("job with path prefix not created")
	}
}

func TestMakePolicy(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakePolicy("Post")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Policies/Post.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("policy not created")
	}

	content, _ := os.ReadFile(path)
	contentStr := string(content)

	// Check for required methods
	if !strings.Contains(contentStr, "type Post struct") {
		t.Error("policy does not contain correct struct")
	}
	expectedMethods := []string{"ViewAny", "View", "Create", "Update", "Delete"}
	for _, method := range expectedMethods {
		if !strings.Contains(contentStr, "func (p *Post) "+method) {
			t.Errorf("policy missing %s method", method)
		}
	}
}

func TestMakePolicyWithPathPrefix(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	err := MakePolicy("Admin/User")
	if err != nil {
		t.Fatal(err)
	}

	path := "app/Policies/Admin/User.go"
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Error("policy with path prefix not created")
	}
}

// Test that nested directories are created properly
func TestMakeGeneratorsCreateNestedDirectories(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	os.Chdir(dir)
	defer os.Chdir(origDir)

	tests := []struct {
		name      string
		generator func(string) error
		path      string
	}{
		{"Controller", func(n string) error { return MakeController(n) }, "app/Http/Controllers/Admin/Api/TestController.go"},
		{"Request", MakeRequest, "app/Http/Requests/Admin/Api/TestRequest.go"},
		{"Command", MakeCommand, "app/Console/Commands/Admin/TestCommand.go"},
		{"Event", MakeEvent, "app/Events/Admin/TestEvent.go"},
		{"Listener", MakeListener, "app/Listeners/Admin/TestListener.go"},
		{"Job", MakeJob, "app/Jobs/Admin/TestJob.go"},
		{"Policy", MakePolicy, "app/Policies/Admin/TestPolicy.go"},
	}

	inputs := []string{
		"Admin/Api/TestController",
		"Admin/Api/TestRequest",
		"Admin/TestCommand",
		"Admin/TestEvent",
		"Admin/TestListener",
		"Admin/TestJob",
		"Admin/TestPolicy",
	}

	for i, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.generator(inputs[i])
			if err != nil {
				t.Fatalf("failed to create %s: %v", tt.name, err)
			}

			if _, err := os.Stat(tt.path); os.IsNotExist(err) {
				t.Errorf("%s with nested path not created at %s", tt.name, tt.path)
			}
		})
	}
}

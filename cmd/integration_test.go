//go:build integration
// +build integration

/*
The integration test workflow:
1. AddTask
2. ListTasks
3. ViewTask
4. CompleteTask
5. ListCompletedTask
6. DeleteTask
7. ListDeletedTask
*/

package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"
)

func randomTaskName(t *testing.T) string {
	t.Helper()

	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	var p strings.Builder

	for i := 0; i < 32; i++ {
		p.WriteByte(chars[r.Intn(len(chars))])
	}

	return p.String()
}

func TestIntegration(t *testing.T) {
	apiRoot := "http://localhost:8080"

	if os.Getenv("TODO_API_ROOT") != "" {
		apiRoot = os.Getenv("TODO_API_ROOT")
	}

	today := time.Now().Format("Jan/02")

	task := randomTaskName(t)
	taskId := ""

	t.Run("AddTask", func(t *testing.T) {
		args := []string{task}
		expOut := fmt.Sprintf("Added task %q to the list.\n", task)

		var out bytes.Buffer

		if err := addAction(&out, apiRoot, args); err != nil {
			t.Fatalf("Expected no error, got %q instead\n", err)
		}

		if expOut != out.String() {
			t.Errorf("Expected output %q, got %q instead\n", expOut, out.String())
		}
	})

	t.Run("ListTask", func(t *testing.T) {
		var out bytes.Buffer

		if err := listAction(&out, apiRoot); err != nil {
			t.Fatalf("Expected no error, got %q instead\n", err)
		}

		outList := ""
		scanner := bufio.NewScanner(&out)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), task) {
				outList = scanner.Text()
				break
			}
		}

		if outList == "" {
			t.Errorf("Task %q is not in the list\n", task)
		}

		taskCompleteStatus := strings.Fields(outList)[0]

		if taskCompleteStatus != "-" {
			t.Errorf("Expected status %q, got %q instead\n", "-", taskCompleteStatus)
		}

		taskId = strings.Fields(outList)[1]
	})

	vRes := t.Run("ViewTask", func(t *testing.T) {
		var out bytes.Buffer

		if err := viewAction(&out, apiRoot, taskId); err != nil {
			t.Fatalf("Expected no error, got %q instead\n", err)
		}

		viewOut := strings.Split(out.String(), "\n")

		if !strings.Contains(viewOut[0], task) {
			t.Fatalf("Expected task %q, got %q instead\n", task, viewOut[0])
		}

		if !strings.Contains(viewOut[1], today) {
			t.Fatalf("Expected creation day/month %q, got %q instead\n", today, viewOut[1])
		}

		if !strings.Contains(viewOut[2], "No") {
			t.Fatalf("Expected completed status %q, got %q instead\n", "No", viewOut[2])
		}
	})

	if !vRes {
		t.Fatalf("View task failed. Stopping integration tests.\n")
	}

	t.Run("CompleteTask", func(t *testing.T) {
		var out bytes.Buffer

		if err := completeAction(&out, apiRoot, taskId); err != nil {
			t.Fatalf("Expected no error, got %q instead\n", err)
		}

		expOut := fmt.Sprintf("Item number %s marked as completed.\n", taskId)

		if expOut != out.String() {
			t.Fatalf("Expected output %q, got %q instead\n", expOut, out.String())
		}
	})

	t.Run("ListCompletedTask", func(t *testing.T) {
		var out bytes.Buffer

		if err := listAction(&out, apiRoot); err != nil {
			t.Fatalf("Expected no error, got %q instead\n", err)
		}

		outList := ""
		scanner := bufio.NewScanner(&out)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), task) {
				outList = scanner.Text()
				break
			}
		}

		if outList == "" {
			t.Errorf("Task %q is not in the list\n", task)
		}

		taskCompleteStatus := strings.Fields(outList)[0]

		if taskCompleteStatus != "X" {
			t.Errorf("Expected status %q, got %q instead\n", "-", taskCompleteStatus)
		}

		taskId = strings.Fields(outList)[1]
	})

	t.Run("DeleteTask", func(t *testing.T) {
		var out bytes.Buffer

		if err := delAction(&out, apiRoot, taskId); err != nil {
			t.Errorf("Expected no error, got %q instead\n", err)
		}

		expOut := fmt.Sprintf("Item number %s deleted.\n", taskId)

		if expOut != out.String() {
			t.Fatalf("Expected output %q, got %q instead\n", expOut, out.String())
		}
	})

	t.Run("ListDeleteAction", func(t *testing.T) {
		var out bytes.Buffer

		if err := listAction(&out, apiRoot); err != nil {
			t.Fatalf("Expected no error, got %q instead\n", err)
		}

		scanner := bufio.NewScanner(&out)
		for scanner.Scan() {
			if strings.Contains(scanner.Text(), task) {
				t.Errorf("Task %q is still in the list\n", task)
				break
			}
		}
	})
}

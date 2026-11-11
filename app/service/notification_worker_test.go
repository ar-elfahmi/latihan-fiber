package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/hibiken/asynq"

	"latihan-fiber/app/model"
	"latihan-fiber/app/repository"
)

// fakeWorkerUsers adalah fake repository.UserRepository untuk NotificationWorker.
type fakeWorkerUsers struct {
	repository.UserRepository // method lain tidak dipakai

	users map[int]model.User
	err   error
}

func (f *fakeWorkerUsers) FindByID(_ context.Context, id int) (model.User, error) {
	if f.err != nil {
		return model.User{}, f.err
	}
	u, ok := f.users[id]
	if !ok {
		return model.User{}, repository.ErrNotFound
	}
	return u, nil
}

var _ repository.UserRepository = (*fakeWorkerUsers)(nil)

// workerTask membuat task Asynq dengan payload job tertentu.
func workerTask(t *testing.T, job model.NotificationJob) *asynq.Task {
	t.Helper()

	payload, err := json.Marshal(job)
	if err != nil {
		t.Fatalf("marshal job: %v", err)
	}
	return asynq.NewTask(model.TaskSendNotification, payload)
}

// TestNotificationWorkerSkipsRetryOnBrokenTaskPayload memastikan payload task
// yang rusak menghasilkan error yang DILETAKKAN dengan asynq.SkipRetry.
//
// Deteksi WAJIB memakai errors.Is: error dibungkus dengan fmt.Errorf("%w: ..."),
// jadi perbandingan "err != asynq.SkipRetry" salah (temuan B1 #3).
func TestNotificationWorkerSkipsRetryOnBrokenTaskPayload(t *testing.T) {
	t.Parallel()

	w := NotificationWorker{Users: &fakeWorkerUsers{users: map[int]model.User{}}}
	task := asynq.NewTask(model.TaskSendNotification, []byte("{bukan json"))

	err := w.HandleSend(context.Background(), task)
	if err == nil {
		t.Fatal("payload rusak harus menghasilkan error")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("error = %v, mau error yang membungkus asynq.SkipRetry", err)
	}
}

// TestNotificationWorkerSkipsRetryOnBrokenEventPayload memastikan payload event
// yang rusak juga dianggap permanen (SkipRetry).
func TestNotificationWorkerSkipsRetryOnBrokenEventPayload(t *testing.T) {
	t.Parallel()

	w := NotificationWorker{Users: &fakeWorkerUsers{users: map[int]model.User{}}}
	// Payload task valid JSON, tapi isi payload event bukan objek JSON,
	// jadi json.Unmarshal ke model.EnrollmentEvent gagal.
	task := asynq.NewTask(model.TaskSendNotification,
		[]byte(`{"outbox_id":1,"event_type":"enrollment.created","payload":"bukan-json"}`))

	err := w.HandleSend(context.Background(), task)
	if err == nil {
		t.Fatal("event rusak harus menghasilkan error")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("error = %v, mau error yang membungkus asynq.SkipRetry", err)
	}
}

// TestNotificationWorkerSkipsRetryOnUnknownUser memastikan user yang tidak ada
// menghasilkan error permanen (SkipRetry).
func TestNotificationWorkerSkipsRetryOnUnknownUser(t *testing.T) {
	t.Parallel()

	w := NotificationWorker{Users: &fakeWorkerUsers{users: map[int]model.User{}}}
	event, err := json.Marshal(model.EnrollmentEvent{UserID: 404, CourseID: 1, CourseCode: "IF101"})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	task := workerTask(t, model.NotificationJob{
		OutboxID:  2,
		EventType: model.EventEnrollmentCreated,
		Payload:   event,
	})

	err = w.HandleSend(context.Background(), task)
	if err == nil {
		t.Fatal("user tak dikenal harus menghasilkan error")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("error = %v, mau error yang membungkus asynq.SkipRetry", err)
	}
}

// TestNotificationWorkerSkipsRetryOnUnknownEvent memastikan event yang tidak
// dikenal menghasilkan error permanen (SkipRetry).
func TestNotificationWorkerSkipsRetryOnUnknownEvent(t *testing.T) {
	t.Parallel()

	w := NotificationWorker{Users: &fakeWorkerUsers{users: map[int]model.User{}}}
	event, err := json.Marshal(model.EnrollmentEvent{UserID: 1, CourseID: 1})
	if err != nil {
		t.Fatalf("marshal event: %v", err)
	}
	task := workerTask(t, model.NotificationJob{
		OutboxID:  3,
		EventType: "event.tidak.dikenal",
		Payload:   event,
	})

	err = w.HandleSend(context.Background(), task)
	if err == nil {
		t.Fatal("event asing harus menghasilkan error")
	}
	if !errors.Is(err, asynq.SkipRetry) {
		t.Fatalf("error = %v, mau error yang membungkus asynq.SkipRetry", err)
	}
}

// TestNotificationWorkerBuildsMessage memastikan pesan yang dibuat worker
// memakai BuildMessage (subjek + penerima sesuai event).
func TestNotificationWorkerBuildsMessage(t *testing.T) {
	t.Parallel()

	user := model.User{ID: 5, Username: "sari", Email: "sari@example.com"}
	event := model.EnrollmentEvent{
		UserID:     5,
		CourseID:   2,
		CourseCode: "IF102",
		CourseName: "Basis Data Lanjut",
	}

	msg, err := BuildMessage(model.EventEnrollmentCreated, event, user)
	if err != nil {
		t.Fatalf("BuildMessage: %v", err)
	}
	if msg.To != user.Email {
		t.Errorf("To = %q, mau %q", msg.To, user.Email)
	}
	if msg.Subject == "" || msg.Body == "" {
		t.Errorf("pesan kosong: %+v", msg)
	}
}

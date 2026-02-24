package mocks

import (
	"time"

	"github.com/stretchr/testify/mock"

	"laporanharianapi/internal/domain"
	"laporanharianapi/internal/repository"
)

// ============================================================
// UserRepositoryMock
// ============================================================

// UserRepositoryMock adalah implementasi mock dari repository.UserRepository.
type UserRepositoryMock struct {
	mock.Mock
}

func (m *UserRepositoryMock) FindByNIP(nip string) (*domain.User, error) {
	args := m.Called(nip)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepositoryMock) FindAll() ([]domain.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *UserRepositoryMock) FindByID(id uint) (*domain.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *UserRepositoryMock) Create(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepositoryMock) Update(user *domain.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *UserRepositoryMock) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *UserRepositoryMock) UpdatePassword(userID uint, newPasswordHash string) error {
	args := m.Called(userID, newPasswordHash)
	return args.Error(0)
}

func (m *UserRepositoryMock) UpdateFoto(userID uint, fotoPath string) error {
	args := m.Called(userID, fotoPath)
	return args.Error(0)
}

func (m *UserRepositoryMock) FindByRoles(roles []string) ([]domain.User, error) {
	args := m.Called(roles)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

func (m *UserRepositoryMock) FindSupervisors(roleFilter string) ([]domain.User, error) {
	args := m.Called(roleFilter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.User), args.Error(1)
}

// ============================================================
// ReportRepositoryMock
// ============================================================

// ReportRepositoryMock adalah implementasi mock dari repository.ReportRepository.
type ReportRepositoryMock struct {
	mock.Mock
}

func (m *ReportRepositoryMock) Create(laporan *domain.Laporan) error {
	args := m.Called(laporan)
	return args.Error(0)
}

func (m *ReportRepositoryMock) CreateFileLaporan(file *domain.FileLaporan) error {
	args := m.Called(file)
	return args.Error(0)
}

func (m *ReportRepositoryMock) CheckIsHoliday(date time.Time) (bool, error) {
	args := m.Called(date)
	return args.Bool(0), args.Error(1)
}

func (m *ReportRepositoryMock) GetAll(filter repository.ReportFilter) ([]domain.Laporan, int64, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.Laporan), args.Get(1).(int64), args.Error(2)
}

// ============================================================
// ReviewRepositoryMock
// ============================================================

// ReviewRepositoryMock adalah implementasi mock dari repository.ReviewRepository.
type ReviewRepositoryMock struct {
	mock.Mock
}

func (m *ReviewRepositoryMock) Create(review *domain.Penilaian) error {
	args := m.Called(review)
	return args.Error(0)
}

func (m *ReviewRepositoryMock) FindByUserID(userID int, limit int, offset int) ([]domain.Penilaian, int64, error) {
	args := m.Called(userID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.Penilaian), args.Get(1).(int64), args.Error(2)
}

func (m *ReviewRepositoryMock) FindByPenilaiID(penilaiID int) ([]domain.Penilaian, error) {
	args := m.Called(penilaiID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Penilaian), args.Error(1)
}

// ============================================================
// TaskRepositoryMock
// ============================================================

// TaskRepositoryMock adalah implementasi mock dari repository.TaskRepository.
type TaskRepositoryMock struct {
	mock.Mock
}

func (m *TaskRepositoryMock) Create(task *domain.TugasPokok) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *TaskRepositoryMock) FindByUserID(userID int) ([]domain.TugasPokok, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TugasPokok), args.Error(1)
}

func (m *TaskRepositoryMock) FindAll() ([]domain.TugasPokok, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TugasPokok), args.Error(1)
}

func (m *TaskRepositoryMock) FindByID(id uint) (*domain.TugasPokok, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TugasPokok), args.Error(1)
}

func (m *TaskRepositoryMock) Update(task *domain.TugasPokok) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *TaskRepositoryMock) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// ============================================================
// NotificationRepositoryMock
// ============================================================

// NotificationRepositoryMock adalah implementasi mock dari repository.NotificationRepository.
type NotificationRepositoryMock struct {
	mock.Mock
}

func (m *NotificationRepositoryMock) Create(notif *domain.Notification) error {
	args := m.Called(notif)
	return args.Error(0)
}

func (m *NotificationRepositoryMock) FindByUserID(userID int) ([]domain.Notification, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Notification), args.Error(1)
}

func (m *NotificationRepositoryMock) MarkAsRead(notifID int, userID int) error {
	args := m.Called(notifID, userID)
	return args.Error(0)
}

// ============================================================
// DashboardRepositoryMock
// ============================================================

// DashboardRepositoryMock adalah implementasi mock dari repository.DashboardRepository.
type DashboardRepositoryMock struct {
	mock.Mock
}

func (m *DashboardRepositoryMock) CountLaporanByUserAndMonth(userID uint, year int, month int) (int64, error) {
	args := m.Called(userID, year, month)
	return args.Get(0).(int64), args.Error(1)
}

func (m *DashboardRepositoryMock) CountTugasPokokByUser(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *DashboardRepositoryMock) CountLaporanHariIni() (int64, error) {
	args := m.Called()
	return args.Get(0).(int64), args.Error(1)
}

func (m *DashboardRepositoryMock) CountLaporanHariIniByRole(role string) (int64, error) {
	args := m.Called(role)
	return args.Get(0).(int64), args.Error(1)
}

func (m *DashboardRepositoryMock) CountTugasPendingHariIni(userID uint) (int64, error) {
	args := m.Called(userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *DashboardRepositoryMock) GetRecentLaporan(userID uint, limit int) ([]domain.Laporan, error) {
	args := m.Called(userID, limit)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Laporan), args.Error(1)
}

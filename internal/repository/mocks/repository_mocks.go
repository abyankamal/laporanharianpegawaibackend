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

func (m *UserRepositoryMock) DeleteWithCleanup(id uint) ([]string, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

func (m *UserRepositoryMock) UpdatePassword(userID uint, newPasswordHash string) error {
	args := m.Called(userID, newPasswordHash)
	return args.Error(0)
}

func (m *UserRepositoryMock) UpdateFoto(userID uint, fotoPath string) error {
	args := m.Called(userID, fotoPath)
	return args.Error(0)
}

func (m *UserRepositoryMock) UpdateFCMToken(userID uint, token string) error {
	args := m.Called(userID, token)
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

func (m *ReportRepositoryMock) GetAll(filter repository.ReportFilter) ([]domain.Laporan, int64, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Get(1).(int64), args.Error(2)
	}
	return args.Get(0).([]domain.Laporan), args.Get(1).(int64), args.Error(2)
}

func (m *ReportRepositoryMock) GetByID(id uint) (*domain.Laporan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Laporan), args.Error(1)
}

func (m *ReportRepositoryMock) GetFileByReportID(reportID uint) (*domain.FileLaporan, error) {
	args := m.Called(reportID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FileLaporan), args.Error(1)
}

func (m *ReportRepositoryMock) GetReportRecap(userID uint, startDate time.Time, endDate time.Time) (*repository.ReportRecapResponse, error) {
	args := m.Called(userID, startDate, endDate)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ReportRecapResponse), args.Error(1)
}

func (m *ReportRepositoryMock) Update(laporan *domain.Laporan) error {
	args := m.Called(laporan)
	return args.Error(0)
}

func (m *ReportRepositoryMock) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *ReportRepositoryMock) GetReportRecapAggregated(filter repository.ReportFilter) (*repository.ReportRecapResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.ReportRecapResponse), args.Error(1)
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

func (m *ReviewRepositoryMock) CheckExistingReview(userID uint, bulan int, tahun int) (bool, error) {
	args := m.Called(userID, bulan, tahun)
	return args.Bool(0), args.Error(1)
}

// ============================================================
// TaskRepositoryMock
// ============================================================

// TaskRepositoryMock adalah implementasi mock dari repository.TaskRepository.
type TaskRepositoryMock struct {
	mock.Mock
}

func (m *TaskRepositoryMock) Create(task *domain.TugasOrganisasi) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *TaskRepositoryMock) FindByAssigneeID(userID int) ([]domain.TugasOrganisasi, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TugasOrganisasi), args.Error(1)
}

func (m *TaskRepositoryMock) FindAll() ([]domain.TugasOrganisasi, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.TugasOrganisasi), args.Error(1)
}

func (m *TaskRepositoryMock) FindByID(id uint) (*domain.TugasOrganisasi, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TugasOrganisasi), args.Error(1)
}

func (m *TaskRepositoryMock) Update(task *domain.TugasOrganisasi) error {
	args := m.Called(task)
	return args.Error(0)
}

func (m *TaskRepositoryMock) ReplaceAssignees(taskID uint, users []domain.User) error {
	args := m.Called(taskID, users)
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

func (m *NotificationRepositoryMock) FindByID(id int, userID int) (*domain.Notification, error) {
	args := m.Called(id, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Notification), args.Error(1)
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

func (m *DashboardRepositoryMock) CountTugasOrganisasiByUser(userID uint) (int64, error) {
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

// ============================================================
// WorkHourRepositoryMock
// ============================================================

type WorkHourRepositoryMock struct {
	mock.Mock
}

func (m *WorkHourRepositoryMock) Get() (*domain.WorkHour, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.WorkHour), args.Error(1)
}

func (m *WorkHourRepositoryMock) Update(workHour *domain.WorkHour) error {
	args := m.Called(workHour)
	return args.Error(0)
}

func (m *WorkHourRepositoryMock) SeedDefault() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================================
// HolidayRepositoryMock
// ============================================================

type HolidayRepositoryMock struct {
	mock.Mock
}

func (m *HolidayRepositoryMock) GetAll() ([]domain.Holiday, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Holiday), args.Error(1)
}

func (m *HolidayRepositoryMock) Create(holiday *domain.Holiday) error {
	args := m.Called(holiday)
	return args.Error(0)
}

func (m *HolidayRepositoryMock) GetByID(id uint) (*domain.Holiday, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Holiday), args.Error(1)
}

func (m *HolidayRepositoryMock) Update(holiday *domain.Holiday) error {
	args := m.Called(holiday)
	return args.Error(0)
}

func (m *HolidayRepositoryMock) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *HolidayRepositoryMock) CheckIsHoliday(date time.Time) (bool, error) {
	args := m.Called(date)
	return args.Bool(0), args.Error(1)
}

// ============================================================
// AbsensiRepositoryMock
// ============================================================

type AbsensiRepositoryMock struct {
	mock.Mock
}

func (m *AbsensiRepositoryMock) Create(absensi *domain.Absensi) error {
	args := m.Called(absensi)
	return args.Error(0)
}

func (m *AbsensiRepositoryMock) Update(absensi *domain.Absensi) error {
	args := m.Called(absensi)
	return args.Error(0)
}

func (m *AbsensiRepositoryMock) GetByUserAndDate(userID uint, tanggal time.Time) (*domain.Absensi, error) {
	args := m.Called(userID, tanggal)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Absensi), args.Error(1)
}

func (m *AbsensiRepositoryMock) GetByUserAndMonth(userID uint, bulan int, tahun int) ([]domain.Absensi, error) {
	args := m.Called(userID, bulan, tahun)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Absensi), args.Error(1)
}

func (m *AbsensiRepositoryMock) GetAllByMonth(bulan int, tahun int) ([]domain.Absensi, error) {
	args := m.Called(bulan, tahun)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Absensi), args.Error(1)
}

func (m *AbsensiRepositoryMock) GetTodayAbsensi(userID uint) (*domain.Absensi, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Absensi), args.Error(1)
}

func (m *AbsensiRepositoryMock) GetAbsensiRecap(userID uint, bulan int, tahun int) (*repository.AbsensiRecapResponse, error) {
	args := m.Called(userID, bulan, tahun)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AbsensiRecapResponse), args.Error(1)
}

// ============================================================
// IzinRepositoryMock
// ============================================================

type IzinRepositoryMock struct {
	mock.Mock
}

func (m *IzinRepositoryMock) Create(izin *domain.PengajuanIzin) error {
	args := m.Called(izin)
	return args.Error(0)
}

func (m *IzinRepositoryMock) Update(izin *domain.PengajuanIzin) error {
	args := m.Called(izin)
	return args.Error(0)
}

func (m *IzinRepositoryMock) GetByID(id uint) (*domain.PengajuanIzin, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.PengajuanIzin), args.Error(1)
}

func (m *IzinRepositoryMock) GetByUserID(userID uint) ([]domain.PengajuanIzin, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PengajuanIzin), args.Error(1)
}

func (m *IzinRepositoryMock) GetPendingApprovals() ([]domain.PengajuanIzin, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PengajuanIzin), args.Error(1)
}

func (m *IzinRepositoryMock) GetApprovedByUserAndDateRange(userID uint, start, end time.Time) ([]domain.PengajuanIzin, error) {
	args := m.Called(userID, start, end)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.PengajuanIzin), args.Error(1)
}

// ============================================================
// SupervisorRepositoryMock
// ============================================================

type SupervisorRepositoryMock struct {
	mock.Mock
}

func (m *SupervisorRepositoryMock) GetSupervisor() (*domain.LurahSupervisor, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.LurahSupervisor), args.Error(1)
}

func (m *SupervisorRepositoryMock) UpdateSupervisor(nama, nip string) error {
	args := m.Called(nama, nip)
	return args.Error(0)
}

func (m *SupervisorRepositoryMock) SeedDefault() error {
	args := m.Called()
	return args.Error(0)
}

// ============================================================
// AdminRepositoryMock
// ============================================================

type AdminRepositoryMock struct {
	mock.Mock
}

func (m *AdminRepositoryMock) GetRekapLaporanAdmin(filter repository.AdminReportFilter) (*repository.AdminReportResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AdminReportResponse), args.Error(1)
}

func (m *AdminRepositoryMock) GetLaporanExportAdmin(filter repository.AdminReportFilter) ([]domain.Laporan, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.Laporan), args.Error(1)
}

func (m *AdminRepositoryMock) GetDashboardSummaryAdmin() (*repository.DashboardSummaryResponse, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.DashboardSummaryResponse), args.Error(1)
}

func (m *AdminRepositoryMock) GetPegawaiAdmin(filter repository.AdminPegawaiFilter) (*repository.AdminPegawaiResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AdminPegawaiResponse), args.Error(1)
}

func (m *AdminRepositoryMock) GetPegawaiStatistik() (*repository.PegawaiStatistik, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PegawaiStatistik), args.Error(1)
}

func (m *AdminRepositoryMock) GetPengumumanAdmin(filter repository.AdminPengumumanFilter) (*repository.AdminPengumumanResponse, error) {
	args := m.Called(filter)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.AdminPengumumanResponse), args.Error(1)
}

func (m *AdminRepositoryMock) GetPengumumanStatistikAdmin() (*repository.PengumumanStatistik, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.PengumumanStatistik), args.Error(1)
}

func (m *AdminRepositoryMock) CreatePengumumanAdmin(pengumuman *domain.Notification) error {
	args := m.Called(pengumuman)
	return args.Error(0)
}

func (m *AdminRepositoryMock) UpdatePengumumanAdmin(id uint, pengumuman *domain.Notification) error {
	args := m.Called(id, pengumuman)
	return args.Error(0)
}

func (m *AdminRepositoryMock) DeletePengumumanAdmin(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

// ============================================================
// JabatanRepositoryMock
// ============================================================

type JabatanRepositoryMock struct {
	mock.Mock
}

func (m *JabatanRepositoryMock) GetAll() ([]domain.RefJabatan, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]domain.RefJabatan), args.Error(1)
}

func (m *JabatanRepositoryMock) GetByID(id uint) (*domain.RefJabatan, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.RefJabatan), args.Error(1)
}

func (m *JabatanRepositoryMock) Create(jabatan *domain.RefJabatan) error {
	args := m.Called(jabatan)
	return args.Error(0)
}

func (m *JabatanRepositoryMock) Update(jabatan *domain.RefJabatan) error {
	args := m.Called(jabatan)
	return args.Error(0)
}

func (m *JabatanRepositoryMock) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}




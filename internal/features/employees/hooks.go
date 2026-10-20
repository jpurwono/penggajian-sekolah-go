package employees

import "context"

// PostCreateHook dipanggil setelah pegawai berhasil dibuat (mis. untuk
// menerapkan template komponen gaji secara otomatis). Kegagalan hook
// bersifat best-effort dan tidak menggagalkan pembuatan pegawai.
type PostCreateHook func(ctx context.Context, employee *Employee) error

var postCreateHook PostCreateHook

// SetPostCreateHook mendaftarkan satu callback post-create (dipanggil dari router).
func SetPostCreateHook(fn PostCreateHook) { postCreateHook = fn }

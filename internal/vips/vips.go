package vips

/*
#cgo pkg-config: vips
#cgo CFLAGS: -O3
#include "vips.h"
*/
import "C"
import (
	"errors"
	"fmt"
	"runtime"
)

var (
	vipsSupportSmartcrop bool
	vipsTypeSupportLoad  = make(map[ImageType]bool)
	vipsTypeSupportSave  = make(map[ImageType]bool)
)

type SubsampleMode int

const (
	VipsForeignSubsampleAuto SubsampleMode = C.VIPS_FOREIGN_JPEG_SUBSAMPLE_AUTO
	VipsForeignSubsampleOn   SubsampleMode = C.VIPS_FOREIGN_JPEG_SUBSAMPLE_ON
	VipsForeignSubsampleOff  SubsampleMode = C.VIPS_FOREIGN_JPEG_SUBSAMPLE_OFF
	VipsForeignSubsampleLast SubsampleMode = C.VIPS_FOREIGN_JPEG_SUBSAMPLE_LAST
)

func Startup() error {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := C.vips_initialize(); err != 0 {
		C.vips_shutdown()
		return fmt.Errorf("unable to start vips!")
	}

	C.vips_cache_set_max_mem(0)
	C.vips_cache_set_max(0)
	C.vips_concurrency_set(1)
	C.vips_vector_set_enabled(0)

	vipsSupportSmartcrop = C.vips_support_smartcrop() == 1

	for _, imgtype := range ImageTypes {
		vipsTypeSupportLoad[imgtype] = int(C.vips_type_find_load_go(C.int(imgtype))) != 0
		vipsTypeSupportSave[imgtype] = int(C.vips_type_find_save_go(C.int(imgtype))) != 0
	}

	return nil
}

func Shutdown() {
	C.vips_shutdown()
}

func vipsGetMem() float64 {
	return float64(C.vips_tracked_get_mem())
}

func vipsGetMemHighwater() float64 {
	return float64(C.vips_tracked_get_mem_highwater())
}

func vipsGetAllocs() float64 {
	return float64(C.vips_tracked_get_allocs())
}

func Cleanup() {
	C.vips_cleanup()
}

func vipsError() error {
	return errors.New(C.GoString(C.vips_error_buffer()))
}

func gbool(b bool) C.gboolean {
	if b {
		return C.gboolean(1)
	}
	return C.gboolean(0)
}

func ibool(b bool) C.int {
	if b {
		return C.int(1)
	}
	return C.int(0)
}

func SupportsWebpAnimation() bool {
	return C.vips_support_webp_animation() != 0
}

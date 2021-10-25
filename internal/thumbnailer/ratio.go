package thumbnailer

import "errors"

func calculateWHWithAspectRatio(w, h, tw, th uint, precise bool) (rw, rh uint, err error) {

	if w == 0 {
		err = errors.New("width not specified")
		return
	}
	if h == 0 {
		err = errors.New("height not specified")
		return
	}
	if tw == 0 {
		err = errors.New("target width not specified")
		return
	}
	if th == 0 {
		err = errors.New("target height not specified")
		return
	}

	rw = tw
	rh = th
	ch := (precise && float64(tw)/float64(th) > float64(w)/float64(h)) ||
		(!precise && float64(w)/float64(tw) > float64(h)/float64(th))
	if ch {
		rh = uint(int(float64(h*tw) / float64(w)))
	} else {
		rw = uint(int(float64(w*th) / float64(h)))
	}
	return
}

package main

// TODO: docs
func (f *Fedora) WithExecScripts(scripts []*File, prePackages bool) *Fedora {
	if prePackages {
		f.ExecScriptPre = append(f.ExecScriptPre, scripts...)
	} else {
		f.ExecScriptPost = append(f.ExecScriptPost, scripts...)
	}

	return f
}

// TODO: docs
func (f *Fedora) WithExec(cmd []string, prePackages bool) *Fedora {
	if prePackages {
		f.ExecPre = append(f.ExecPre, cmd)
	} else {
		f.ExecPost = append(f.ExecPost, cmd)
	}

	return f
}

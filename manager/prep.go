package manager

func (ps *DownloadStep) Run(task *Task) error {
	_, err := task.DownloadFile(ps.Source)
	return err
}

/*
func (ps *CopyFileStep) Run(task *Task) error {
	if ps.ArgsCopy != "" {
		dstPath := path.Join(task.Workdir, ps.ArgsCopy.Dest)
		srcPath := task.Inputs[ps.ArgsCopy.Source]
		log.Printf("Copy %s to %s", srcPath, dstPath)
		cpCmd := exec.Command("cp", "-rf", srcPath, dstPath)
		err := cpCmd.Run()
		return err
	}
}
*/

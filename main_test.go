package main

import "testing"

func Test_processResult(t *testing.T) {

	testConfigFIle := getInitialTestData()

	type args struct {
		filesToProcess []InternalAppFile
	}
	tests := []struct {
		testName string
		args     args
		init     func()
		want     func() bool
	}{
		/*
			before:
			-------------------
			src:
				a:
					aa.txt
					ab.txt
				b:
					aa.txt
					ab.txt
			-------------------
			after:
			-------------------
			src:
				a:
					aa.txt [DELETED]
					ab.txt
				b:
					aa.txt [DELETED]
					ab.txt
			-------------------

		*/
		{
			testName: "Delete removes only desired files",
			args: args{
				filesToProcess: []InternalAppFile{},
			},
		},
		/*
			before:
			-------------------
			src:
				a:
					aa.txt
					ab.txt
				b:
					aa.txt
					ab.txt
			dest:

			-------------------
			after:
			-------------------
			src:
				a:
					aa.txt
					ab.txt
					c.txt [EXISTS]
				b:
					aa.txt
					ab.txt
			dest:
				c.txt [EXISTS]
			-------------------
		*/
		{
			testName: "Copy doesn't remove file",
			args: args{
				filesToProcess: []InternalAppFile{},
			},
		},
		/*
			before:
			-------------------
			src:
				a:
					aa.txt
					ab.txt
				b:
					aa.txt
					ab.txt
			dest:

			-------------------
			after:
			-------------------
			src:
				a:
					aa.txt
					ab.txt
				b:
					aa.txt
					ab.txt
			dest:
				aa.txt
				aa.1.txt
			-------------------
		*/
		{
			testName: "Copy resolves conflicts",
			args: args{
				filesToProcess: []InternalAppFile{},
			},
		},
		{
			testName: "Copy doesn't remove file (preserve relative path option)",
			args: args{
				filesToProcess: []InternalAppFile{},
			},
		},
		{
			testName: "Move copy file and remove old one",
			args: args{
				filesToProcess: []InternalAppFile{},
			},
		},
		{
			testName: "Move copy file and remove old one (preserve relative path option)",
			args: args{
				filesToProcess: []InternalAppFile{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			preparePlayground()

			processResult(tt.args.filesToProcess)

			cleanPlayground()
		})
	}
}

func preparePlayground() {
	testConfigFIle := parseUserInput()
}

func cleanPlayground() {

}

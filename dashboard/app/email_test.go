// Copyright 2017 syzkaller project authors. All rights reserved.
// Use of this source code is governed by Apache 2 LICENSE that can be found in the LICENSE file.

// +build aetest

package dash

import (
	"testing"

	"github.com/google/syzkaller/pkg/email"
)

func TestEmailReport(t *testing.T) {
	c := NewCtx(t)
	defer c.Close()

	build := testBuild(1)
	c.expectOK(c.API(client2, key2, "upload_build", build, nil))

	crash := testCrash(build, 1)
	crash.Maintainers = []string{`"Foo Bar" <foo@bar.com>`, `bar@foo.com`}
	c.expectOK(c.API(client2, key2, "report_crash", crash, nil))

	c.expectOK(c.GET("/email_poll"))
	c.expectEQ(len(c.emailSink), 1)
	msg := <-c.emailSink
	sender, _, err := email.RemoveAddrContext(msg.Sender)
	if err != nil {
		t.Fatalf("failed to remove sender context: %v", err)
	}
	c.expectEQ(sender, fromAddr(c.ctx))
	to := config.Namespaces["test2"].Reporting[0].Config.(*EmailConfig).Email
	c.expectEQ(msg.To, []string{to})
	c.expectEQ(msg.Subject, crash.Title)
	c.expectEQ(len(msg.Attachments), 1)
	c.expectEQ(msg.Attachments[0].Name, "config.txt")
	c.expectEQ(msg.Attachments[0].Data, build.KernelConfig)
	body := `
Hello,syzkaller hit the following crash on kernel_commit1
repo1/branch1
compiler: compiler1
.config is attached


CC: ["Foo Bar" <foo@bar.com> bar@foo.com]

report1

---
This bug is generated by a dumb bot. It may contain errors.
Direct all questions to syzkaller@googlegroups.com.

`
	c.expectEQ(msg.Body, body)
}
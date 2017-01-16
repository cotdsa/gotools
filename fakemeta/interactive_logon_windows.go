package main

import (
	"fmt"
	"github.com/mash/go-tempfile-suffix"
	"os"
	"os/exec"
	"strings"
)

func get_user_password_interactive() (string, string, error) {
	var (
		cmdOut []byte
		err    error
	)
	cmmm := `
[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Drawing")
[void] [System.Reflection.Assembly]::LoadWithPartialName("System.Windows.Forms")

$objForm = New-Object System.Windows.Forms.Form 
$objForm.Text = "Enter Credentials"
$objForm.Size = New-Object System.Drawing.Size(300,200) 
$objForm.StartPosition = "CenterScreen"

$objForm.KeyPreview = $True
$objForm.Add_KeyDown({if ($_.KeyCode -eq "Enter") 
    {$user=$objTextBox.Text; $pass=$objTextBoxx.Text;$objForm.Close()}})
$objForm.Add_KeyDown({if ($_.KeyCode -eq "Escape") 
    {$objForm.Close()}})

$OKButton = New-Object System.Windows.Forms.Button
$OKButton.Location = New-Object System.Drawing.Size(75,130)
$OKButton.Size = New-Object System.Drawing.Size(75,23)
$OKButton.Text = "OK"
$OKButton.Add_Click({$user=$objTextBox.Text; $pass=$objTextBoxx.Text; $objForm.Close()})
$objForm.Controls.Add($OKButton)

$CancelButton = New-Object System.Windows.Forms.Button
$CancelButton.Location = New-Object System.Drawing.Size(150,130)
$CancelButton.Size = New-Object System.Drawing.Size(75,23)
$CancelButton.Text = "Cancel"
$CancelButton.Add_Click({$objForm.Close()})
$objForm.Controls.Add($CancelButton)

$objLabel = New-Object System.Windows.Forms.Label
$objLabel.Location = New-Object System.Drawing.Size(10,20) 
$objLabel.Size = New-Object System.Drawing.Size(280,20) 
$objLabel.Text = "Username"
$objForm.Controls.Add($objLabel) 

$objTextBox = New-Object System.Windows.Forms.TextBox 
$objTextBox.Location = New-Object System.Drawing.Size(10,40) 
$objTextBox.Size = New-Object System.Drawing.Size(260,20) 
$objForm.Controls.Add($objTextBox) 

$objLabel2 = New-Object System.Windows.Forms.Label
$objLabel2.Location = New-Object System.Drawing.Size(10,70) 
$objLabel2.Size = New-Object System.Drawing.Size(280,20) 
$objLabel2.Text = "Password"
$objForm.Controls.Add($objLabel2) 

$objTextBoxx = New-Object System.Windows.Forms.TextBox 
$objTextBoxx.Location = New-Object System.Drawing.Size(10,90) 
$objTextBoxx.Size = New-Object System.Drawing.Size(260,20) 
$objTextBoxx.PasswordChar = "*"
$objForm.Controls.Add($objTextBoxx) 

$objForm.Topmost = $True

$objForm.Add_Shown({$objForm.Activate()})
[void] $objForm.ShowDialog()

$user
$pass
`

	tempFile, err := tempfile.TempFileWithSuffix(os.TempDir(), "myTempFile", ".ps1")
	defer os.Remove(tempFile.Name())

	if err != nil {
		return "", "", err
	}

	tempFile.WriteString(strings.Replace(cmmm, "\n", "\r\n", -1))
	tempFile.Close()
	args := []string{"-WindowStyle", "hidden", "-NoProfile", "-NonInteractive", "-ExecutionPolicy", "Unrestricted", "-File", tempFile.Name()}
	if cmdOut, err = exec.Command("PowerShell", args...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	var outputdata []string = strings.Split(string(cmdOut), "\r\n")

	if len(outputdata) == 3 && outputdata[0] != "" && outputdata[1] != "" {
		return outputdata[0], outputdata[1], nil
	} else {
		return "", "", fmt.Errorf("Data not returned")
	}
}

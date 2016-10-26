# How to contribute

## Steps to follow :scroll:

### 1. Fork it :fork_and_knife:

You can get your own fork/copy of [tiberious](https://github.com/JustAnotherOrganization/tiberious) by using the <a href="https://github.com/JustAnotherOrganization/tiberious/new/master?readme=1#fork-destination-box"><kbd><b>Fork</b></kbd></a> button or clicking [this](https://github.com/JustAnotherOrganization/tiberious/new/master?readme=1#fork-destination-box).

 [![Fork Button](https://help.github.com/assets/images/help/repository/fork_button.jpg)](https://github.com/JustAnotherOrganization/tiberious)

### 2. Clone it :busts_in_silhouette:

You need to clone (download) it to local machine using

```sh
$ git clone https://github.com/Your_Username/tiberious.git
```

> This makes a local copy of repository in your machine.

Once you have cloned the `tiberious` repository in Github, move to that folder first using change directory command on linux and Mac.

```sh
# This will change directory to a folder tiberious
$ cd tiberious
```

Move to this folder for all other commands.

### 3. Set it up :arrow_up:

Run the following commands to see that *your local copy* has a reference to *your forked remote repository* in Github :octocat:

```sh
$ git remote -v
origin  https://github.com/Your_Username/tiberious.git (fetch)
origin  https://github.com/Your_Username/tiberious.git (push)
```

Now, lets add a reference to the original [tiberious](https://github.com/JustAnotherOrganization/tiberious) repository using

```sh
$ git remote add upstream https://github.com/JustAnotherOrganization/tiberious.git
```

> This adds a new remote named ***upstream***.

See the changes using

```sh
$ git remote -v
origin    https://github.com/Your_Username/tiberious.git (fetch)
origin    https://github.com/Your_Username/tiberious.git (push)
upstream  https://github.com/JustAnotherOrganization/tiberious.git (fetch)
upstream  https://github.com/JustAnotherOrganization/tiberious.git (push)
```

### 4. Sync it :recycle:

Always keep your local copy of repository updated with the original repository.
Before making any changes and/or in an appropriate interval, run the following commands *carefully* to update your local repository.

```sh
# Fetch all remote repositories and delete any deleted remote branches
$ git fetch --all --prune

# Switch to `master` branch
$ git checkout master

# Reset local `master` branch to match `upstream` repository's `master` branch
$ git reset --hard upstream/master

# Push changes to your forked `tiberious` repo
$ git push origin master
```

### 5. Adding and updating dependencies

Install godep
```sh
$ go get github.com/tools/godep
```
Use godep to get all of the dependencies
```sh
godep restore
```

### 6. Ready Steady Go... :turtle: :rabbit2:

Once you have completed these steps, you are ready to start contributing by checking our Issues and creating [pull requests](https://github.com/JustAnotherOrganization/tiberious/pulls).

### 7. Create a new branch :bangbang:

Whenever you are going to make contribution. Please create seperate branch using command and keep your `master` branch clean (i.e. synced with remote branch).

```sh
# It will create a new branch with name Branch_Name and switch to branch Folder_Name
$ git checkout -b Folder_Name
```

Create a seperate branch for contibution and try to use same name of branch as of folder.

To switch to desired branch

```sh
# To switch from one folder to other
$ git checkout Folder_Name
```

To add the changes to the branch. Use

```sh
# To add all files to branch Folder_Name
$ git add .
```

Type in a message relevant for the code reveiwer using

```sh
# This message get associated with all files you have changed
$ git commit -m 'relevant message'
```

Now, Push your awesome work to your remote repository using

```sh
# To push your work to your remote repository
$ git push -u origin Folder_Name
```

Finally, go to your repository in browser and click on `compare and pull requests`.
Then add a title and description to your pull request that explains your precious effort.

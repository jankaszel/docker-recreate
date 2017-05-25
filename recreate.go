package main

import (
  "fmt"
  "os"
  "strconv"
  "strings"
  "time"

  "github.com/fsouza/go-dockerclient"
  //"github.com/tonnerre/golang-pretty"
)

func checkError(err error) {
  if err != nil {
    fmt.Println(err)
    os.Exit(0)
  }
}

func main() {
  if len(os.Args) < 2 {
    fmt.Printf("Usage: %s [-p] id [tag]\n", os.Args[0])
    os.Exit(0)
  }

  client, err := docker.NewClientFromEnv()
  checkError(err)

  args, err := parseArgs(os.Args)

  recentContainer, err := client.InspectContainer(args.containerId)
  checkError(err)

  // TODO delete _new if an error occures

  repository, currentTag := parseImageName(recentContainer.Config.Image)

  if args.tagName == "" {
    args.tagName = currentTag
  }

  fmt.Printf("Image: %s:%s\n", repository, args.tagName)

  if args.pullImage {
    fmt.Print("Pulling image...\n")

    err = client.PullImage(docker.PullImageOptions{
      Repository: repository,
      Tag: args.tagName }, docker.AuthConfiguration{})

    checkError(err)
  }

  // TODO handle image tags/labels?

  now := int(time.Now().Unix())
  then := now - 1

  name := recentContainer.Name
  temporaryName := name + "_" + strconv.Itoa(now)
  recentName := name + "_" + strconv.Itoa(then)

  // TODO possibility to add/change environment variables
  var options docker.CreateContainerOptions
  options.Name = temporaryName
  options.Config = recentContainer.Config
  options.Config.Image = repository + ":" + args.tagName
  options.HostConfig = recentContainer.HostConfig
  options.HostConfig.VolumesFrom = []string{recentContainer.ID}

  links := recentContainer.HostConfig.Links

  for i := range links {
    parts := strings.SplitN(links[i], ":", 2)
    if len(parts) != 2 {
      fmt.Println("Unable to parse link ", links[i])
      // TODO make function and add better error return
      return
    }

    containerName := strings.TrimPrefix(parts[0], "/")
    aliasParts := strings.Split(parts[1], "/")
    alias := aliasParts[len(aliasParts)-1]
    links[i] = fmt.Sprintf("%s:%s", containerName, alias)
  }
  options.HostConfig.Links = links

  fmt.Println("Creating...")
  newContainer, err := client.CreateContainer(options)
  checkError(err)

  err = client.RenameContainer(docker.RenameContainerOptions{
    ID: recentContainer.ID,
    Name: recentName })
  checkError(err)

  err = client.RenameContainer(docker.RenameContainerOptions{
    ID: newContainer.ID,
    Name: name})
  checkError(err)

  if recentContainer.State.Running {
    fmt.Printf("Stopping old container\n")
    err = client.StopContainer(recentContainer.ID, 10)
    checkError(err)

    fmt.Printf("Starting new container\n")
    err = client.StartContainer(newContainer.ID, newContainer.HostConfig)
    checkError(err)
  }

  // TODO fallback to old container if error occured

  if args.deleteContainer {
    fmt.Printf("Deleting old container...\n")

    err = client.RemoveContainer(docker.RemoveContainerOptions{
      ID: recentContainer.ID,
      RemoveVolumes: false })
    checkError(err)
  }

  fmt.Printf(
    "Migrated from %s to %s\n",
    recentContainer.ID[:4],
    newContainer.ID[:4])

  fmt.Println("Done")
}

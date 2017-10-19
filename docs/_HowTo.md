# How To

## How to Update Your Application on Kelda
The most robust way to handle updates to your application is to [build and
push](https://docs.docker.com/get-started/part2/) your own tagged Docker images.
We recommend always [tagging images](https://docs.docker.com/engine/reference/commandline/build/#tag-an-image--t),
and not using the `:latest` tag.

Say that we want to update an application that uses the `me/myWebsite:0.1`
Docker image to use `me/myWebsite:0.2` instead. We can do that in two simple
steps:

1. In the blueprint, update all references to `me/myWebsite:0.1` to use the tag
`:0.2`.
2. Run `kelda run` with the updated blueprint to update the containers with the
new image.

Kelda will now restart all the relevant containers to use the new tag.

### Untagged Images and Images Specified in Blueprints
For users who do not use tagged Docker images, there are currently two ways of
updating an application. This section explains the two methods and when to use
each.

#### After Changing a Blueprint
If changes are made directly to the blueprint or a file or module that the
blueprint depends on, simply `kelda run` the blueprint again. Kelda will detect
the changes and reboot the affected containers.

Examples of when `kelda run` will update the application:

* You updated the contents of a file that is transferred to the application
container using `withFiles`.
* You changed the Dockerfile string passed to the `Image` constructor in the
blueprint.


#### Updating an Image
Though we recommend using _tagged_ Docker images, some applications might use
untagged images either hosted in a registry like Docker Hub or created with
Kelda's `Image` constructor in a blueprint. To pull a newer version of a hosted
image or rebuild an `Image` object with Kelda, you need to restart the relevant
container:

1. In the blueprint, remove the code that `.deploy()`s the container that
should be updated.
2. `kelda run` this modified blueprint in order to stop the container.
3. Back in the blueprint, add the `.deploy()` call back in to the blueprint.
4. `kelda run` the blueprint from step 3 to reboot the container with the
updated image.

Examples of when you need to stop and start the container in order to update the
application:

* You use a hosted image (e.g. on Docker Hub), and pushed a new image with the
same `name:tag` as the currently running one.
* You want to rebuild an `Image` even though the Dockerfile in the blueprint
didn't change. For instance, this could be because the Dockerfile clones a
GitHub repository, so rebuilding the image would recreate the container with
the latest code from GitHub.

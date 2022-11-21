![screenshot of overview page](./demo/overview.png)
![screenshot of details page](./demo/details.png)
![demo video](https://github.com/codesoap/mycolog/releases/download/v0.1.0/demo.mp4)

mycolog helps you keep an overview of your mushroom cultivation
projects. It can store notes for each component, so that it's easier to
remember which experiments succeeded and which failed. Genetics can be
traced through family trees.

# Installation
You can download the precompiled program from the
[releases page](https://github.com/codesoap/mycolog/releases).

In order to get family trees displayed, you need to have Graphviz
installed. You can download it [here](https://graphviz.org/download/).

If you want to compile the program yourself, do this:

```bash
git clone git@github.com:codesoap/mycolog.git
cd mycolog
go install
# The binary is now at ~/go/bin/mycolog.
```

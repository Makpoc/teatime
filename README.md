teatime
=======

A command line timer for tea brewing with small eye candies

Installation 
============
Make sure you have a working golang setup and use `go get`.

```
go get github.com/makpoc/teatime
```

Usage
=====
```
teatime -help
Usage of teatime:
  -duration="":     Tee timer duration. Can be Xs/m/h (overwrite -tea's default duration if given) or +-Xs/m/h (add to it)
  -file="":         Path to json file, containing tea specifications
  -list=false:      List all available tea types and exit with brief information about each of them
  -tea="":          Type of Tea to prepare (either the Name or the ID. See -list)
```
#### Options:
`-tea` - select the tea by name or ID and start the timer

`-duration` - must be in `time.Duration` format (*180s* or *3m*). Can also be prefixed by - or + sign, which will add to or remove from the base duration of a selected `-tea`. For example the following if "Green Dragon" needs 120 seconds we can add some more time by executing `teatime -tea "Green Dragon" -duration +30s`. Negative time is not allowed.

`-list` of available teas. They are taken from `-file` if specified or from a predefined (default) set.

`-file` expects a json file with "list of teas". If not given - a default set is used. The json contains a list of objects with the following properties:

* **id** - for convenience when selecting tea - this can be used instead of the name.
* **type** - the tea type. e.g. Green, Black. Used for information only.
* **name** - the name of the tea.
* **steepTime** - needs a time unit, parsable by `time.Duration` (*180s* or *3m*). This is used as base duration.
* **temp** - the recommended water temperature. Used for information only.

##### Note
Notifications are provided by [go-notify](https://github.com/mqu/go-notify "go-notify"). On my system it's using notify-osd, which was showing the pop up on a very awkward place (*almost* top-right on second monitor). See [here](http://askubuntu.com/questions/128474/how-to-customize-on-screen-notifications "Ask Ubuntu") for details on how to move the dialog.

Examples
=======
List all available teas from a file:

```
$ teatime -file sample/teas.json -list

>       Tea Time(r)
>          ____    ,-^-,
>       ,|'----'|  * L *
>      ((|      |  '-.-'
>       \|      |
>        |      |
>        '------'
>      ^^^^^^^^^^^^
> Id:         0
> Name:       White Dragon
> Type:       White
> Steep Time: 2 minutes
> Tempreture: 70°
> ------
> Id:         1
> Name:       Temple of Heaven
> Type:       Green
> Steep Time: 3 minutes
> Tempreture: 80°
> ------
> [...]
```
Start a timer for and add 1 minute to the base duration:
```
$ teatime -tea "Temple of Heaven" -duration +1m

> Id:         1
> Name:       Temple of Heaven
> Type:       Green
> Steep Time: 2 minutes
> Tempreture: 80°
> 
>       Tea Time(r)
>          ____    ,-^-,
>       ,|'----'|  * L *
>      ((|      |  '-.-'
>       \|      |
>        |      |
>        '------'
>      ^^^^^^^^^^^^
> Progress: [##########] (100%) |   0/180 seconds remaining
> Your tea is ready! Enjoy :)
```

License
=======
Use/copy/modify as you see fit :)

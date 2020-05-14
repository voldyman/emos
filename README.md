#### emos - Emoji Search

I wrote this program because emoji search on https://discordemojis.com is pretty slow and i want to be annoying in our work chat more often.

The CLI caches the emojis and builds an index to make it faster to search.


##### Usage

```
$ emos '*snug'
MonkaSnug - https://discordemoji.com/assets/emoji/MonkaSnug.png

$ emos -link r
https://discordemoji.com/assets/emoji/r.png

$ emos -md pepehug
pepehug - /md ![](https://discordemoji.com/assets/emoji/pepehug.png)

$ emos -md -link -lucky monkaHmm
/md ![](https://discordemoji.com/assets/emoji/pepehug.png)

```

My usual usage is 

```
$ emos -md -link -lucky cry | pbcopy
```



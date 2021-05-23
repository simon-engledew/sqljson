Transform a mysqldump into a json file, the JSON file into a dot file, and the dot file into a ERD.

```
mysqldump database | sqljsondump | sqljsondot | dot -Tpng -o database.png && open database.png
```

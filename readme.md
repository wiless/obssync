# Download your notebook with all attachments
./obssync -url <url> -o mynotebook.obs  -d 


# Run the file 
dataflow run mynotebook.ojs --allow-file-attachments 


# NOTE that observablehq uses latest version of plot library, you can add the Plot requirement in the standard library

create `custom.js` with the following content
```
window.DATAFLOW_STDLIB = {
  constants: {
    name: "alex",
  },
  dependency: {
    require: {
      // moment: (require) => require("moment"),
      Plot: (require) => require("@observablehq/plot"),
    },
  },
};
```


# Run with custom library 
dataflow run mynotebook.ojs --allow-file-attachments --stdlib=custom.js

document$.subscribe(function() {
  var tables = document.querySelectorAll("article table:not([class])")
  tables.forEach(function(table) {
    new Tabulator(table, {
      columns:[
        {title:"Group", field:"group", width:150, headerFilter:"input"},
        {title:"Version", field:"version", width:100, headerFilter:"input"},
        {title:"Kind", field:"kind", width:250, headerFilter:"input"},
        {title:"Deprecated", field:"deprecated", width:120, headerFilter:"input"},
        {title:"Deleted", field:"deleted", width:120, headerFilter:"input"},
        {title:"Replacement", field:"replacement", width:400, headerFilter:"input"},
      ]
    })
  })
})
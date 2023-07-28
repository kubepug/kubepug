new gridjs.Grid({
  columns: [
  {
    name: "Group",
    id: "group",
    sort: true,
    width: "10%",
  },
  {
    name: "Version",
    id: "version",
    sort: true,
    width: "5%",
  },
  {
    name: "Kind",
    id: "kind",
    sort: true,
    width: "10%",
  },
  {
    name: "Description",
    id: "description",
    sort: false,
    width: "25%",
  },
  {
    name: "Deprecated",
    data: (row) => "v" + row.deprecated_version.version_major + "." + row.deprecated_version.version_minor,
    sort: true,
    width: "5%",
  },
  {
    name: "Deleted",
    data: (row) => "v" + row.removed_version.version_major + "." + row.removed_version.version_minor,
    sort: true,
    width: "5%",
  },
  {
    name: "Replacement",
    data: (row) => {
                if (row.replacement.group != null && row.replacement.group != undefined) {
                    return  gridjs.html("<b>Group: </b>" + row.replacement.group + "<br>" +
                            "<b>Version: </b>" + row.replacement.version + "<br>" +
                            "<b>Kind: </b>" + row.replacement.kind)
                }
    },
    sort: false,
  }
  ],
  sort: true,
  pagination: false,
  fixedHeader: true,
  search: true,
  server: {
    url: '/data/data.json',
    then: data => data,
  }
}).render(document.getElementById("wrapper"));

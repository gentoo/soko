var useflagChartCreated = false;

// wait for d3 and render the bubb
var checkD3 = setInterval(function() {
    if (typeof d3 !== 'undefined') {
        clearInterval(checkD3);

        if(!useflagChartCreated){
            createUseflagChart();
        }

        useflagChartCreated = true;
    }
}, 100);

function createUseflagChart() {
    
    if(!useflagChartCreated) {

        $('#bubble-placeholder').show();

        var width = 600;
        height = 600;

        var diameter = 960,
            format = d3.format(",d"),
            color = d3.scale.category20c();

        var bubble = d3.layout.pack()
            .sort(null)
            .size([width, height])
            .padding(1.5);

        var svg = d3.select("#bubble-placeholder").append("svg")
            .attr("width", width)
            .attr("height", height)
            .attr("class", "bubble");

        d3.json("/useflags/popular.json", function (error, root) {
            if (error) throw error;

            var node = svg.selectAll(".node")
                .data(bubble.nodes(classes(root))
                    .filter(function (d) {
                        return !d.children;
                    }))
                .enter().append("g")
                .attr("class", "node")
                .attr("transform", function (d) {
                    return "translate(" + d.x + "," + d.y + ")";
                });

            node.append("title")
                .text(function (d) {
                    return d.className + ": " + format(d.value);
                });

            node.append("circle")
                .attr("r", function (d) {
                    return d.r;
                })
                .attr("class", "kk-useflag-circle")
                .attr("onclick", function (d) {
                    return "location.href='/useflags/" + d.className + "';";
                })
                .style("fill", function (d) {
                    return color(d.className);
                });

            node.append("text")
                .attr("dy", ".3em")
                .attr('class', 'kk-useflag-circle')
                .attr("onclick", function (d) {
                    return "location.href='/useflags/" + d.className + "';";
                })
                .style("text-anchor", "middle")
                .style("font-size", function (d) {
                    var len = d.className.substring(0, d.r / 3).length;
                    var size = d.r / 3;
                    size *= 8 / len;
                    if (len == 1) {
                        size -= 60;
                    }
                    size += 1;
                    return Math.round(size) + 'px';
                })
                .text(function (d) {
                    return d.className.substring(0, d.r / 3);
                });
        });

        // Returns a flattened hierarchy containing all leaf nodes under the root.
        function classes(root) {
            var classes = [];

            function recurse(name, node) {
                if (node.children) node.children.forEach(function (child) {
                    recurse(node.name, child);
                });
                else classes.push({packageName: name, className: node.name, value: node.size});
            }

            recurse(null, root);
            return {children: classes};
        }

        d3.select(self.frameElement).style("height", height + "px");
    }
}

import { select as d3Select } from 'd3-selection';
import { json as d3Json } from 'd3-fetch';
import { format as d3Format } from 'd3-format';
import { scaleOrdinal as d3ScaleOrdinal } from 'd3-scale';
import { pack as d3Pack, hierarchy as d3Hierarchy } from 'd3-hierarchy';

const schemeCategory20c = [
    "#3182bd", "#6baed6", "#9ecae1", "#c6dbef",
    "#e6550d", "#fd8d3c", "#fdae6b", "#fdd0a2",
    "#31a354", "#74c476", "#a1d99b", "#c7e9c0",
    "#756bb1", "#9e9ac8", "#bcbddc", "#dadaeb",
    "#636363", "#969696", "#bdbdbd", "#d9d9d9",
];

var useflagChartCreated = false;

// wait for d3 and render the bubb
var checkD3 = setInterval(function () {
    clearInterval(checkD3);

    if (!useflagChartCreated) {
        createUseflagChart();
    }

    useflagChartCreated = true;
}, 100);

// TODO this is a workaround for now
// as we do get duplicate charts from
// time to time. This is probably related
// to turbolinks.
function deleteDuplicate() {
    var bubbles = document.querySelectorAll(".bubble");
    while (bubbles.length > 1) {
        bubbles[1].remove();
        bubbles = document.querySelectorAll(".bubble");
    }
}

setTimeout(deleteDuplicate, 100);
setTimeout(deleteDuplicate, 200);
setTimeout(deleteDuplicate, 500);
setTimeout(deleteDuplicate, 1000);

function createUseflagChart() {
    if (!useflagChartCreated) {
        $('#bubble-placeholder').show();

        const width = 600;
        const height = 600;

        const fmt = d3Format(",d"),
            color = d3ScaleOrdinal(schemeCategory20c);

        const bubble = d3Pack()
            .size([width, height])
            .padding(1.5);

        const svg = d3Select("#bubble-placeholder").append("svg")
            .attr("width", width)
            .attr("height", height)
            .attr("class", "bubble");

        d3Json("/useflags/popular.json").then(function (root) {
            const useflags = root.map(function (d) {
                return { packageName: "flags", className: d.name, value: d.size };
            });
            const hier = d3Hierarchy({ children: useflags });
            hier.sum(function (d) { return d.value; });

            const nodes = bubble(hier).leaves();

            const node = svg.selectAll(".node")
                .data(nodes)
                .enter().append("g")
                .attr("class", "node")
                .attr("transform", function (d) { return "translate(" + d.x + "," + d.y + ")"; });

            node.append("title")
                .text(function (d) { return d.data.className + ": " + fmt(d.value); });

            node.append("circle")
                .attr("r", function (d) { return d.r; })
                .attr("class", "kk-useflag-circle")
                .attr("onclick", function (d) { return "location.href='/useflags/" + d.data.className + "';"; })
                .style("fill", function (d) { return color(d.data.className); });

            node.append("text")
                .attr("dy", ".3em")
                .attr('class', 'kk-useflag-circle')
                .attr("onclick", function (d) { return "location.href='/useflags/" + d.data.className + "';"; })
                .style("text-anchor", "middle")
                .style("font-size", function (d) {
                    var len = d.data.className.substring(0, d.r / 3).length;
                    var size = d.r / 3;
                    size *= 8 / len;
                    if (len == 1) {
                        size -= 60;
                    }
                    size += 1;
                    return Math.round(size) + 'px';
                })
                .text(function (d) { return d.data.className.substring(0, d.r / 3); });
            d3Select(self.frameElement).style("height", height + "px");
        });
    }
}

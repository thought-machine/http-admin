/* global $ */
/* global google */

function ChartRenderer(element, title) {
  const header = $('<h5></h5>')
    .addClass('text-center')
    .text(title);
  const chartDiv = $('<div></div>')[0];

  $(element)
    .empty()
    .append(header);
  $(element).append(chartDiv);

  this.columns = {};

  this.chart = new google.visualization.LineChart(chartDiv);
  this.chartData = new google.visualization.DataTable();
  this.chartData.addColumn('datetime', 'Time');
  this.chartData.addRows(60);
  this.chartOptions = {
    legend: {position: 'none'},
    width: '100%',
    height: 350,
    chartArea: {width: '90%', height: '85%'},
    vAxis: {
      viewWindow: {min: 0},
      baselineColor: '#ddd',
      gridlineColor: '#ddd',
    },
    hAxis: {
      baselineColor: '#ddd',
      gridlineColor: '#ddd',
      textStyle: {
        fontSize: 12,
      },
    },
  };

  this.chart.draw(this.chartData, this.chartOptions);
}

ChartRenderer.prototype.appendMetric = function(metrics) {
  const row = [new Date()];
  const c = this;
  metrics.forEach(function(metric) {
    if (!(metric.name in c.columns)) {
      c.columns[metric.name] = true;
      c.chartData.addColumn('number', metric.name);
    }
    row.push(metric.value);
  });
  this.addRow(row);
};

ChartRenderer.prototype.addRow = function(row) {
  this.chartData.addRow(row);
  this.chartData.removeRow(0);
  this.chart.draw(this.chartData, this.chartOptions);
};

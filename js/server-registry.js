/* global $ */
/* global google */
/* global ChartRenderer */

google.load('visualization', '1', {packages: ['corechart'], callback: graphLibLoaded});

function graphLibLoaded() {
  let interval = {};

  function refreshStats(dds, chartRenderer) {
    clearInterval(interval);
    let url = $('#server-tabs').data('refresh-uri') + '?';
    let label;
    for (let i = 0; i < dds.length; i++) {
      const key = $(dds[i]).data('key');
      url += 'm=' + key + '&';
      label = $(dds[i]).data('label');
    }

    function render(data) {
      const json = $.parseJSON(data);
      let failures = 0;
      let requests = 0;
      for (let i = 0; i < json.length; i++) {
        const name = json[i].name;
        const value = json[i].value;
        const id = name.replace(/\/|{|}|=/g, '-');
        if (name.indexOf(label) > -1) {
          if (name.indexOf('failures') > -1) failures = value;
          else if (name.indexOf('requests') > -1) requests = value;
        }
        $('#' + id).text(value);
      }

      let sr = 0.0;
      if (requests > 0) sr = Number(((1.0 - failures / requests) * 100.0).toFixed(4));
      chartRenderer.appendMetric([{name: '', value: sr}]);
    }

    interval = setInterval(function() {
      $.ajax({
        url,
        dataType: 'text',
        success: render,
      });
    }, 1000);
  }

  $('a[data-toggle="tab"]').on('shown.bs.tab', function() {
    const active = $('#servers').find('.tab-pane.active');
    const graphDiv = active.find('#server-graph');
    const chart = new ChartRenderer(graphDiv[0], 'Success Rate');
    refreshStats(active.find('dd'), chart);
  });

  $('#server-tabs a:first').tab('show');
}

/* global $ */
/* global google */
/* global ChartRenderer */

google.load('visualization', '1', {packages: ['corechart'], callback: graphLibLoaded});

function graphLibLoaded() {
  const charDiv = $('#chart-div')[0];
  let selected = undefined;
  let interval = {};

  function refreshStats(stat, chartRenderer) {
    clearInterval(interval);
    const url = $('#metrics-grid').data('refresh-uri') + '?m=' + stat;

    function render(data) {
      const json = $.parseJSON(data);
      if (json[0] !== undefined) chartRenderer.appendMetric(json);
    }

    interval = setInterval(function() {
      $.ajax({
        url,
        dataType: 'text',
        success: render,
      });
    }, 1000);
  }

  function render(li) {
    const stat = li.html();
    if (selected !== undefined) selected.removeClass('selected');
    li.addClass('selected');
    selected = li;
    refreshStats(stat, new ChartRenderer(charDiv, stat));
  }

  $('#metrics li').on('click', function(e) {
    render($(e.target));
  });

  const fragmentId = $(
    '#' +
      window.location.hash
        .replace('#', '')
        .replace(/\//g, '-')
        // css chars to escape: !"#$%&'()*+,-./:;<=>?@[\]^`{|}~
        .replace(/(!|"|#|%|&|'|\(|\)|\*|\+|,|-|\.|\/|:|;|<|=|>|\?|@|\[|\\|\]|\^|`|{|\||}|~)/g, '\\$1')
  );
  if (fragmentId[0] !== undefined) {
    $(fragmentId)[0].scrollIntoView(true);
    render(fragmentId);
  } else {
    render($('#metrics li:first'));
  }
}

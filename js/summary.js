/* global $ */
/* global MsToStringConverter */
/* global BytesToStringConverter */

const waitForDom = setInterval(function() {
  if ($('#process-info') !== null) {
    clearInterval(waitForDom);
    loadProcInfo();
    loadClientInfo();
    loadServerInfo();
    loadLintInfo();
  }
}, 250);

function loadProcInfo() {
  let url = $('#process-info').data('refresh-uri') + '?';
  const list = $('#process-info ul li');

  for (let i = 0; i < list.length; i++) {
    const key = $(list[i]).data('key');
    if (key !== undefined) url += '&m=' + key;
  }

  const msToStr = new MsToStringConverter();
  const bytesToStr = new BytesToStringConverter();

  function pretty(name, value) {
    if (name === 'process_uptime') return msToStr.convert(value * 1000);
    else if (name === 'go_memstats_alloc_bytes') return bytesToStr.convert(value);
    else if (name === 'go_gc_duration_seconds') return msToStr.convert(value * 1000);
    return value;
  }

  function renderProcInfo(data) {
    const json = $.parseJSON(data);
    for (let i = 0; i < json.length; i++) {
      const id = json[i].name.replace(/\//g, '-');
      const value = pretty(json[i].name, json[i].value);
      $('#' + id).text(value);
    }
  }

  function fetchProcInfo() {
    $.ajax({
      url,
      dataType: 'text',
      cache: false,
      success: renderProcInfo,
    });
  }

  fetchProcInfo();
  setInterval(fetchProcInfo, 1000);
}

function loadClientInfo() {
  function fetchClientInfo() {
    $.ajax({
      url: $('#client-info').data('refresh-uri'),
      dataType: 'text',
      cache: false,
      success(data) {
        $('#client-info').html(data);
      },
    });
  }

  fetchClientInfo();
  setInterval(fetchClientInfo, 1000);
}

function loadServerInfo() {
  $.ajax({
    url: $('#server-info').data('refresh-uri'),
    dataType: 'text',
    cache: false,
    success(data) {
      $('#server-info').html(data);
    },
  });
}

function loadLintInfo() {
  $.ajax({
    url: $('#lint-warnings').data('refresh-uri') + '?',
    dataType: 'text',
    cache: false,
    success(data) {
      $('#lint-warnings').html(data);
    },
  });
}

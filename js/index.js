/* global $ */

$(document).ready(function() {
  $('#toggle').click(function(e) {
    e.preventDefault();
    $('#wrapper').toggleClass('toggled');
    $('#toggle span').toggleClass('fa-angle-right');
  });

  $('nav .subnav-title').click(function() {
    $(this)
      .parent()
      .find('ul')
      .slideToggle();
    $(this)
      .find('.fas')
      .toggleClass('fa-caret-square-up');
  });

  const confirmationLinks = {
    'Abort-Server': 'abort',
    'Quit-Server': 'quit',
    Shutdown: 'shutdown',
  };

  for (const key in confirmationLinks) {
    if (confirmationLinks.hasOwnProperty(key)) {
      const elem = document.getElementById(key);
      $(elem).click({name: confirmationLinks[key]}, function(event) {
        return confirm('Are you sure you want to ' + event.data.name + '?');
      });
    }
  }
});

$(function() {
  editor = window.editor;

  var datasetId = 0;
  var old_program = '';

  var compile_spinner = $('.compile-spinner');
  compile_spinner.hide();
  var compile_button = $('.compile-button');


  window.show_attempt = function(msg) {
    window.location.href = window.domainname + '/attempt/' + msg.id;
  };

  window.submit_code = function() {
    var program = editor.getValue();
    if (datasetId == undefined) {
      datasetId = 0;
    }
    /*
    window.setTimeout(
      function() {
        window.location.href = pathname + '/attempts'
      },
      200000
    );
    */
    begin_spinner(compile_spinner, compile_button);
    $.ajax({
      type: 'POST',
      async: true,
      url: pathname + '/submit',
      data: {
        'program': program,
        'datasetId': datasetId
      },
      timeout: 360000,
      dataType: 'json',
      success: function(msg) {
        if (msg.status == 'success') {
          var req = new Pollymer.Request({
            maxDelay: 120000,
            timeout: 120000
          });
          req.on('finished', function(code, result, headers) {
            if (result.status == 'success') {
              show_attempt(result);
            } else if (result.status == 'error-not-found') {
              req.retry();
            } else {
              req.abort();
              show_error(result);
              end_spinner(compile_spinner, compile_button);
            }
          });
          req.on('error', function(reason) {
            var result = {
              title: 'Error: Was not able to compile program',
              message: 'Was not able to compile the program. ' +
              'The system might be under a heavy load.'
            };
            show_error(result);
            end_spinner(compile_spinner, compile_button);
          });
          req.maxTries = 30;
          req.start('GET', pathname + '/attemptrun/' + msg.runId);
        } else {
          show_error(msg);
          end_spinner(compile_spinner, compile_button);
        }
      }
    });
  };


  window.save_answer = function(number, answer) {
    if (number == undefined || answer == undefined) {
      return;
    }
    $.ajax({
      type: 'POST',
      async: false,
      url: pathname + '/question/save/' + number,
      data: {
        'answer': answer
      },
      timeout: 10000,
      dataType: 'json',
      success: function(msg) {
        if (msg.status == 'success') {
          $.bootstrapGrowl('Answer to question ' + number + ' saved.', {
            type: 'success',
            delay: 1200
          });
        } else if (msg.status == 'error-deadline') {
          var str = 'Error: Was not able to save answer for question ' +
                    number + ' because deadline for this MP has passed.';
          $.bootstrapGrowl(str, {
            type: 'danger',
            delay: 2000
          });
        } else {
          var str = 'Error: Was not able to save answer for question ' +
                    number + '. ';
          if (msg.data != undefined) {
            str += msg.data;
          }
          $.bootstrapGrowl(str, {
            type: 'danger',
            delay: 2000
          });
        }
      },
      error: function() {
        $.bootstrapGrowl('Error: Was not able to save answer.', {
          type: 'danger',
          delay: 3000
        });
      }
    });
  };

  $('.save-answer')
    .click(function() {
      var number = $(this)[0].id;
      var answer = $('#' + number + '.answer').val();
      save_answer(number, answer);
    });

  $('.dataset-choice')
    .click(function() {
      datasetId = parseInt($(this)[0].id);
      submit_code();
    });

  var update_history = function(id, program, date) {
    var html =
      '<tr>' +
      '<td>' +
      $.trim(program)
      .slice(0, 50) +
      '<a href=\'/program/' + id + '\'>' +
      '&hellip;' +
      '</a>' +
      '</td>' +
      '<td>' +
      $.timeago(date) +
      '</td>' +
      '</tr>';
    $('tbody#program-history')
      .prepend(html);
  };

  /*
  var old_program = editor.getValue();
  var program_save_request = new Pollymer.Request();
  program_save_request.on('finished', function(code, result, headers) {
    var program = editor.getValue();
    program_save_request.SkipSend = program == old_program;
    console.log(program_save_request.SkipSend);
    if (!program_save_request.SkipSend) {
      old_program = program;
    }
    program_save_request._body = {
      'program': program
    };
    program_save_request.retry();
  });
  program_save_request.off('error');
  program_save_request.maxTries = -1;
  program_save_request.maxDelay = 5000;
  program_save_request.start('POST', pathname + '/save', {}, {
      'program': old_program
  });
  */

  // auto save
  var program = editor.getValue();
  var num_save_failed = 0;
  (function poll(interval0) {
    var interval = interval0 || 5000;
    setTimeout(
      function() {
        if (program === old_program) {
          return poll(interval);
        }
        $.ajax({
          type: 'POST',
          async: true,
          url: pathname + '/save',
          data: {
            'program': program
          },
          timeout: interval,
          dataType: 'json',
          success: function(msg) {
            if (msg.status == 'success') {
              old_program = program;
              $.bootstrapGrowl('Program saved.', {
                type: 'success',
                delay: 1200
              });
              update_history(msg.id, program, Date.now());
            } else {
              $.bootstrapGrowl('Error: Was not able to save program.', {
                type: 'danger',
                delay: 600
              });
            }
          },
          error: function() {
            $.bootstrapGrowl('Error: Was not able to save program.', {
              type: 'danger',
              delay: 600
            });
            if (num_save_failed++ > 10) {
              interval += 1000;
            } else {
              return;
            }
          },
          complete: function() {
            poll(interval);
          }
        });
      },
      interval
    );
  })();

  $(document).on('keydown', function(e) {
    if (e.ctrlKey && e.which === 83) {
        var program = editor.getValue();
        $.ajax({
          type: 'POST',
          async: false,
          url: pathname + '/save',
          data: {
            'program': program
          },
          timeout: 3000,
          dataType: 'json',
          success: function(msg) {
            if (msg.status == 'success') {
              old_program = program;
              $.bootstrapGrowl('Program saved.', {
                type: 'success',
                delay: 1200
              });
              update_history(msg.id, program, Date.now());
            } else {
              $.bootstrapGrowl('Error: Was not able to save program.', {
                type: 'danger',
                delay: 600
              });
            }
          },
          complete: function() {
            old_program = program;
          }
        });
        $('.question-number').html(function(e, id) {
          var number = id.trim();
          var answer = $('#' + number + '.answer').val();
          save_answer(number, answer);
        });
        e.preventDefault();
    }
  });

  if (getURLParameter('show') == 'code') {
    $('#code-tab').tab('show');
  } else if (getURLParameter('show') == 'history') {
    $('#history-tab').tab('show');
  }

});


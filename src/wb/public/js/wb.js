$(function() {

  window.stripTrailingSlash = function(str) {
    if (_.isString(str)) {
      return str.replace(/\/$/, '');
    } else {
      return str;
    }
  };
  if (!window.location.origin) {
    window.location.origin = window.location.protocol +
                             '//' + window.location.hostname +
                             (window.location.port ?
                              ':' + window.location.port : '');
  }
  window.domainname = stripTrailingSlash(window.location.origin);

  window.pathname = stripTrailingSlash(window.location.pathname);

  window.getURLParameter = function(name) {
    name = name.replace(/[\[]/, '\\\[').replace(/[\]]/, '\\\]');
    var regex = new RegExp('[\\?&]' + name + '=([^&#]*)'),
        results = regex.exec(location.search);
    return results == null ? '' :
           decodeURIComponent(results[1].replace(/\+/g, ' '));
  };

  $.timeago.settings.allowFuture = true;

  $('abbr.timeago')
    .timeago();

  window.log_user_data = function() {
    $.ajax({
      type: 'POST',
      async: true,
      url: window.domainname + '/log_page_view',
      timeout: 10000,
      dataType: 'json',
      data: {}
    });
  };

  log_user_data();

  window.begin_spinner = function(elem, btn) {
    btn.hide();
    elem.spin({
      top: 10,
      left: 90,
      length: 5,
      right: 0
    });
    elem.show();
  };

  window.end_spinner = function(elem, btn) {
    elem.hide();
    btn.show();
    elem.spin(false);
  };

  var escapeNewLine = function(s) {
    return s.replace(/(\r\n|\n|\r)/gm, '<br>');
  };

  var takeNLines = function(s, n) {
    return s.split(/(\r\n|\n|\r)/).slice(0, n).join('');
  };

  window.show_error = function(data) {
    if (_.has(data, 'link')) {
      bootbox.dialog({
        message:
          '<pre>' +
            escapeNewLine(takeNLines(data.data || 'Unknown error', 30)) +
          '</pre>',
        title: data.title || 'Error',
        className: 'bootbox-alert',
        buttons: {
          attempt: {
            label: 'Show Attempt',
            className: 'btn-danger',
            callback: function() {
              window.location.href = data.link;
            }
          },
          ok: {
            label: 'Cancel',
            className: 'btn-success',
            callback: function() {
              return;
            }
          }
        }
      });
    } else {
      bootbox.alert({
        title: data.title || 'Error',
        message:
          '<pre>' +
            escapeNewLine(takeNLines(data.data || 'Unknown error', 30)) +
          '</pre>'
      });
    }
  };

  window.show_grade = function(msg) {
    window.location.href = window.domainname + '/grade/' + msg.id;
  };


  $('.submit-grade-spinner').hide();

  window.submit_grade = function() {
    var grade_btn = $('.submit-grade-btn');
    var grade_spinner = $('.submit-grade-spinner');
    var mp_id = $('mp_num').val();
    /*
    window.setTimeout(
      function() {
        window.location.href = window.domainname + '/grades/' + mp_num
      },
      200000
    );
    */
    begin_spinner(grade_spinner, grade_btn);
    $.ajax({
      type: 'POST',
      async: true,
      url: pathname + '/grade',
      timeout: 600000,
      dataType: 'json',
      success: function(msg) {
        if (msg.status == 'success') {
          var req = new Pollymer.Request({
            maxDelay: 120000,
            timeout: 120000
          });
          req.on('finished', function(code, result, headers) {
            if (result.status == 'success') {
              show_grade(result);
            } else if (result.status == 'error-not-found') {
              req.retry();
            } else {
              req.abort();
              show_error(result);
              end_spinner(grade_spinner, grade_btn);
            }
          });
          req.on('error', function(reason) {
            var result = {
              title: 'Error: Was not able to grade program',
              message: 'Was not able to grade the program. ' +
              'The system might be under a heavy load.'
            };
            show_error(result);
            end_spinner(grade_spinner, grade_btn);
          });
          req.maxTries = 30;
          req.start('GET', pathname + '/graderun/' + msg.runId);
        } else {
          end_spinner(grade_spinner, grade_btn);
          show_error(msg);
        }
      }
    });
  };

  window.submit_coursera_grade = function(grade_id, type, forceQ) {
    $.ajax({
      type: 'POST',
      async: false,
      url: domainname + '/coursera/post_grade/' + grade_id +
           '/' + type + '/' + forceQ,
      timeout: 600000,
      success: function(msg) {
        if (msg.status == 'success') {
          window.location.href = window.pathname;
        } else {
          show_error(msg);
        }
      }
    });
  };

  $('.submit-grade-btn')
    .click(function() {
      submit_grade();
    });

  $('.submit-coursera-grade-btn')
    .click(function() {
      submit_coursera_grade($(this).attr('id'), 'all', 'false');
    });

  $('.force-submit-coursera-code-grade')
    .click(function() {
      submit_coursera_grade($(this).attr('id'), 'code', 'true');
    });

  $('.force-submit-coursera-peer-grade')
    .click(function() {
      submit_coursera_grade($(this).attr('id'), 'peer', 'true');
    });

  CodeMirror.addKeywords(
    '__device__ __host__ __global__ dim3 _shared__ __constant__ ' +
    'wbLog OFF FATAL ERROR WARN INFO DEBUG TRACE ' +
    '__syncthreads cudaMemcpyHostToDevice cudaMemcpyDeviceToHost cudaMemcpy ' +
    'cudaFree cudaMalloc cudaThreadSynchronize wbTime_stop wbTime_start ' +
    'wbImport malloc threadIdx blockIdx blockDim '
  );

  CodeMirror.commands.autocomplete = function(cm) {
    return CodeMirror.showHint(cm, CodeMirror.csharpHint);
  };

  $('.code-editor').each(function(ii) {
    var editor = CodeMirror.fromTextArea($(this)[0], {
      lineNumbers: true,
      indentWithTabs: false,
      smartIndent: true,
      tabSize: 4,
      indentUnit: 4,
      indentWithTabs: true,
      readOnly: $(this).hasClass('readonly'),
      theme: 'eclipse',
      mode: 'text/x-c++src',
      matchBrackets: true,
      extraKeys: {
        'Ctrl-Space': 'autocomplete'
      }
    });

    $('#tab-interface a[data-toggle="tab"]')
      .on('shown.bs.tab', function(e) {
        editor.refresh();
      });

    editor.setSize(null, '80%');

    window.editor = editor;
  });

  var collectReviews = function() {
    var reviews = [];
    var score_map = {
      Unsatisfactory: 1,
      Poor: 2,
      Average: 3,
      Good: 4,
      Excellent: 5
    };
    $('.review-tab').each(function(ii) {
      var id = parseInt($(this).attr('id'));
      var question_score =
        score_map[$('#' + id + '.question-review-score').val()];
      var question_comment = $('#' + id + '.question-review-comment').val();
      var code_score = score_map[$('#' + id + '.code-review-score').val()];
      var code_comment = $('#' + id + '.code-review-comment').val();
      var review = {
        id: id,
        question_score: question_score,
        question_comment: $.trim(question_comment),
        code_score: code_score,
        code_comment: $.trim(code_comment)
      };
      reviews.push(review);
    });
    return reviews;
  };

  $('.submit-peer-review')
    .click(function() {
      var mp_id = parseInt($('.mp-id').attr('id'));
      console.log(collectReviews());
      $.ajax({
        type: 'POST',
        async: false,
        url: window.domainname + '/peerreview/' + mp_id,
        timeout: 60000,
        dataType: 'json',
        data: {
          mp_id: mp_id,
          reviews: JSON.stringify(collectReviews())
        },
        success: function(msg) {
          if (msg.status == 'success') {
            bootbox.dialog({
              title: msg.title || 'Peer Rreview Submitted',
              message: msg.data || 'Your peer review has been logged.',
              className: 'bootbox-alert',
              buttons: {
                attempt: {
                  label: 'Show Grade',
                  className: 'btn-danger',
                  callback: function() {
                    window.location.href = msg.link;
                  }
                },
                ok: {
                  label: 'Cancel',
                  className: 'btn-success',
                  callback: function() {
                    return;
                  }
                }
              }
            });
          } else {
            bootbox.alert({
              'title': msg.title || 'Error',
              'message': msg.data || 'Was not able to post peer review.'
            });
          }
        },
        error: function() {
          bootbox.alert({
            'title': 'Error',
            'message': 'Was not able to post peer review.'
          });
        }
      });
    });
});


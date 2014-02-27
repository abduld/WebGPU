$(window).load(function() {
  var imgs = $("img[title='thumbnail']");
  imgs.each(function() {
    var maxWidth = 512;
    var maxHeight = 256;
    var ratio = 0;
    var width = $(this).width();
    var height = $(this).height();
    if (width > maxWidth) {
      ratio = maxWidth / width;
      $(this).css('width', maxWidth);
      $(this).css('heigh', height * ratio);
      height = height * ratio;
    }

    var width = $(this).width();
    var height = $(this).height();
    if (height > maxHeight) {
      ratio = maxHeight / height;
      $(this).css('height', maxHeight);
      $(this).css('width', width * ratio);
      width = width * ratio;
    }
    var alt = $(this).attr('alt');
    var src = $(this).attr('src');
    $(this).wrap(
      '<div class="row-fluid"><div class="span12 centered"></div></div>'
    ).wrap(
      '<a data-lightbox href="' + src + '" class="img-thumbnail"></a>'
    );
  });
});


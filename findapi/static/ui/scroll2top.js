
function enableToTop() {
    // create a button element to allow user to
    // scroll to top of page.
    var $toTop = $('<a>', { href: '#'})
        .addClass('btn btn-scroll2top')
        .css({
            'position': 'fixed',
            'bottom': '25px',
            'right': '25px',
            'display': 'none'
        })
        .append(
            $('<i>').addClass('fas fa-chevron-up')
        )
        .click(function () {
            $('body').animate({scrollTop: 0}, 'slow');
            // $('body').scrollTop('slow');
        });

    $('body').append($toTop);

    // add event listener for page scrolling to
    // control the visibility of button element.
    $(window).scroll(function () {
        if ($(this).scrollTop() > 50) {
            $toTop.fadeIn('fast');
        } else {
            $toTop.fadeOut('fast');
        }
    });
}

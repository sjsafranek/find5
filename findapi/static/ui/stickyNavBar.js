
function enableStickyNavBar() {
    //
    var toggleAffix = function(affixElement, scrollElement) {
        var height = affixElement.outerHeight()
        if (scrollElement.scrollTop() > 10){
            affixElement.addClass("affix");
        } else {
            affixElement.removeClass("affix");
        }
    };

    $('[data-toggle="affix"]').each(function() {
        var $elem = $(this);
        $(window).on('scroll resize', function(e) {
            toggleAffix($elem, $(this));
        });
        toggleAffix($elem, $(window));
    });
}

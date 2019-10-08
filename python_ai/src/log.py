import logging


def NewLogger(name):
    setupLogger(name)
    return logging.getLogger(name)


def setupLogger(name):
    logger = logging.getLogger(name)
    logger.setLevel(logging.DEBUG)
    # fh = logging.FileHandler(name + '.log')
    # fh.setLevel(logging.DEBUG)
    ch = logging.StreamHandler()
    ch.setLevel(logging.DEBUG)
    formatter = logging.Formatter(
        '%(asctime)s - [%(name)s/%(funcName)s] - %(levelname)s - %(message)s')
    # fh.setFormatter(formatter)
    ch.setFormatter(formatter)
    # logger.addHandler(fh)
    logger.addHandler(ch)
    # return logger

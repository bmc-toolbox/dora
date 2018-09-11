Dora
====

Developing locally
------------------

Create a dora.yaml to be used by the application to read settings. This file is placed in:

``~/.bmc-toolbox/dora.yaml``

You can grab valid settings from any bkbuild box:

``sudo cat /etc/bmc-toolbox/dora.yaml``

Start the local webserver:

.. code-block:: bash

    cd ~/go/src/gitlab.booking.com/go/dora
    # Example with debug: true
    $ go run main.go server
    [GIN-debug] [WARNING] Creating an Engine instance with the Logger and Recovery middleware already attached.
    [GIN-debug] [WARNING] Running in "debug" mode. Switch to "release" mode in production.
    - using env:   export GIN_MODE=release
    - using code:  gin.SetMode(gin.ReleaseMode)

    [GIN-debug] GET    /api_static/*filepath     --> gitlab.booking.com/go/dora/vendor/github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)
    [GIN-debug] HEAD   /api_static/*filepath     --> gitlab.booking.com/go/dora/vendor/github.com/gin-gonic/gin.(*RouterGroup).createStaticHandler.func1 (3 handlers)

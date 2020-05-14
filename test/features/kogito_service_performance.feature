# Commented code will be addressed by further enhancements:
# https://issues.redhat.com/browse/KOGITO-1701
# https://issues.redhat.com/browse/KOGITO-1888

@performance
Feature: Kogito Service Performance

  Background:
    Given Namespace is created

  @quarkus
  Scenario Outline: Quarkus Kogito Service Performance with native <native>, without persistence and with requests <requests>
    Given Kogito Operator is deployed
    And Deploy quarkus example service "process-quarkus-example" with configuration:
      | config      | native       | <native> |
      | runtime-env | JAVA_OPTIONS | -Xmx10G  |
    And Kogito application "process-quarkus-example" has 1 pods running within <minutes> minutes
    And Service "process-quarkus-example" with process name "orders" is available within 3 minutes

    When <requests> HTTP POST requests with report using 100 threads on service "process-quarkus-example" with path "orders" and body:
      """json
      {
        "approver" : "john",
        "order" : {
          "orderNumber" : "12345",
          "shipped" : false
        }
      }
      """

    Then Service "process-quarkus-example" contains <requests> instances of process with name "orders" within 1 minutes
    And Service "process-quarkus-example" contains <requests> instances of process with name "orderItems" within 1 minutes
    #And All human tasks on path "orderItems" with path task name "Verify_order" are successfully "completed" with timing "true"


    Examples:
      | native   | minutes | requests |
      | disabled | 10      | 40000    |
      | disabled | 10      | 80000    |
#      | disabled | 10      | 160000   |
#      | disabled | 10      | 320000   |

    @native
    Examples:
      | native  | minutes | requests |
      | enabled | 20      | 40000    |
      | enabled | 20      | 80000    |
#      | enabled | 20      | 160000   |
#      | enabled | 20      | 320000   |

#####

  @quarkus
  @persistence
  Scenario Outline: Quarkus Kogito Service Performance with native <native>, with persistence and with requests <requests>
    Given Kogito Operator is deployed with Infinispan operator
    And Infinispan instance "external-infinispan" is deployed for performance within 5 minute(s) with configuration:
      | username | developer |
      | password | mypass    |
    And Deploy quarkus example service "process-quarkus-example" with configuration:
      | config      | native       | <native>                  |
      | runtime-env | JAVA_OPTIONS | -Xmx10G                   |
      | infinispan  | username     | developer                 |
      | infinispan  | password     | mypass                    |
      | infinispan  | uri          | external-infinispan:11222 |
    And Kogito application "process-quarkus-example" has 1 pods running within <minutes> minutes
    And Service "process-quarkus-example" with process name "orders" is available within 3 minutes

    When <requests> HTTP POST requests with report using 100 threads on service "process-quarkus-example" with path "orders" and body:
      """json
      {
        "approver" : "john",
        "order" : {
          "orderNumber" : "12345",
          "shipped" : false
        }
      }
      """

    #Then Service "process-quarkus-example" contains <requests> instances of process with name "orders" within 1 minutes
    #And Service "process-quarkus-example" contains <requests> instances of process with name "orderItems" within 1 minutes
    #And All human tasks on path "orderItems" with path task name "Verify_order" are successfully "completed" with timing "true"


    Examples:
      | native   | minutes | requests |
      | disabled | 10      | 40000    |
      | disabled | 10      | 80000    |
#      | disabled | 10      | 160000   |
#      | disabled | 10      | 320000   |

    @native
    Examples:
      | native  | minutes | requests |
      | enabled | 20      | 40000    |
      | enabled | 20      | 80000    |
#      | enabled | 20      | 160000   |
#      | enabled | 20      | 320000   |

#####

  @springboot
  Scenario Outline: Spring Boot Kogito Service Performance without persistence and with requests <requests>
    Given Kogito Operator is deployed
    And Deploy springboot example service "process-springboot-example" with configuration:
      | runtime-env | JAVA_OPTIONS | -Xmx10G |
    And Kogito application "process-springboot-example" has 1 pods running within <minutes> minutes
    And Service "process-springboot-example" with process name "orders" is available within 3 minutes

    When <requests> HTTP POST requests with report using 100 threads on service "process-springboot-example" with path "orders" and body:
      """json
      {
        "approver" : "john",
        "order" : {
          "orderNumber" : "12345",
          "shipped" : false
        }
      }
      """

    Then Service "process-springboot-example" contains <requests> instances of process with name "orders" within 1 minutes
    And Service "process-springboot-example" contains <requests> instances of process with name "orderItems" within 1 minutes
    #And All human tasks on path "orderItems" with path task name "Verify_order" are successfully "completed" with timing "true"

    Examples:
      | minutes | requests |
      | 10      | 40000    |
      | 10      | 80000    |
#      | 10      | 160000   |
#      | 10      | 320000   |

#####

  @springboot
  @persistence
  Scenario Outline: Spring Boot Kogito Service Performance with persistence and with requests <requests>
    Given Kogito Operator is deployed with Infinispan operator
    And Infinispan instance "external-infinispan" is deployed for performance within 5 minute(s) with configuration:
      | username | developer |
      | password | mypass    |
    And Deploy springboot example service "process-springboot-example" with configuration:
      | runtime-env | JAVA_OPTIONS | -Xmx10G                   |
      | infinispan  | username     | developer                 |
      | infinispan  | password     | mypass                    |
      | infinispan  | uri          | external-infinispan:11222 |
    And Kogito application "process-springboot-example" has 1 pods running within <minutes> minutes
    And Service "process-springboot-example" with process name "orders" is available within 3 minutes

    When <requests> HTTP POST requests with report using 100 threads on service "process-springboot-example" with path "orders" and body:
      """json
      {
        "approver" : "john",
        "order" : {
          "orderNumber" : "12345",
          "shipped" : false
        }
      }
      """

    #Then Service "process-springboot-example" contains <requests> instances of process with name "orders" within 1 minutes
    #And Service "process-springboot-example" contains <requests> instances of process with name "orderItems" within 1 minutes
    #And All human tasks on path "orderItems" with path task name "Verify_order" are successfully "completed" with timing "true"

    Examples:
      | minutes | requests |
      | 10      | 40000    |
      | 10      | 80000    |
#      | 10      | 160000   |
#      | 10      | 320000   |

<?php

/**
 * Function that does the dirty to connect to the Icinga 2 API
 *
 * @param string $url to connect to.
 *
 * @return array
 */
function connectIcinga2($url) {

  global $curl_stats;

  $state = icinga2v1Get($url, 'objects/hosts');
  $hosts = [];
  if (isset($state['results'])) {
    foreach ($state['results'] as $host) {
      $hosts[$host['name']]                                  = $host['attrs'];
      $hosts[$host['name']]['services']                      = [];
      $hosts[$host['name']]['downtimes']                     = [];
      $hosts[$host['name']]['current_state']                 = $host['attrs']['state'];
      $hosts[$host['name']]['last_state_change']             = $host['attrs']['last_state_change'];
      $hosts[$host['name']]['problem_has_been_acknowledged'] = $host['attr']['acknowledgement'];
      $hosts[$host['name']]['scheduled_downtime_depth']      = $host['attrs']['downtime_depth'];
      $hosts[$host['name']]['notifications_enabled']         = $host['attrs']['enable_notifications'] ? 1 : 0;
      $hosts[$host['name']]['plugin_output']                 = $host['attrs']['output'];
      $hosts[$host['name']]['current_attempt']               = $host['attrs']['check_attempt'];
      $hosts[$host['name']]['max_attempts']                  = $host['attrs']['max_check_attempts'];
      $hosts[$host['name']]['state_type']                    = $host['attrs']['state_type'];
    }
  } else {
    echo $state[0];
  }

  $state = icinga2v1Get($url, 'objects/services');
  if (isset($state['results'])) {
    foreach ($state['results'] as $service) {
      $hn                                                           = $service['attrs']['host_name'];
      $sn                                                           = $service['attrs']['display_name'];
      $hosts[$hn]['services'][$sn]                                  = $service['attrs'];
      $hosts[$hn]['services'][$sn]['downtimes']                     = [];
      $hosts[$hn]['services'][$sn]['current_state']                 = $service['attrs']['state'];
      $hosts[$hn]['services'][$sn]['problem_has_been_acknowledged'] = $service['attrs']['acknowledgement'];
      $hosts[$hn]['services'][$sn]['scheduled_downtime_depth']      = $service['attrs']['downtime_depth'];
      $hosts[$hn]['services'][$sn]['notifications_enabled']         = $service['attrs']['enable_notifications'] ? 1 : 0;
      $hosts[$hn]['services'][$sn]['plugin_output']                 = $service['attrs']['last_check_result']['output'];
      $hosts[$hn]['services'][$sn]['max_attempts']                  = $service['attrs']['max_check_attempts'];
      $hosts[$hn]['services'][$sn]['current_attempt']               = $service['attrs']['check_attempt'];
      $hosts[$hn]['services'][$sn]['state_type']                    = $service['attrs']['state_type'];
    }
  } else {
    echo $state[0];
  }

  $state = icinga2v1Get($url, 'objects/downtimes');
  if (isset($state['results'])) {
    foreach ($state['results'] as $downtime) {
      $hn = $downtime['attrs']['host_name'];
      $sn = $downtime['attrs']['service_name'];
      if ($sn == "") {
        // host downtime
        $hosts[$hn]['downtimes'][] = $downtime;
      } else {
        // service downtime
        $hosts[$hn]['services'][$sn][] = $downtime;
      }
    }
  } else {
    echo $state[0];
  }

  return $hosts;
}

/**
 * Query Icinga 2 API.
 *
 * @param string $url to query.
 * @param string $endpoint to query data from.
 *
 * @return array|mixed
 */
function icinga2v1Get($url, $endpoint) {
  $hostname    = parse_url($url, PHP_URL_HOST);
  $port        = parse_url($url, PHP_URL_PORT);
  $username    = parse_url($url, PHP_URL_USER);
  $password    = parse_url($url, PHP_URL_PASS);
  $request_url = "$url/v1/{$endpoint}";
  $headers     = [
    'Accept: application/json',
    'X-HTTP-Method-Override: GET',
  ];
  $ch          = curl_init();
  curl_setopt_array(
    $ch, [
      CURLOPT_URL            => $request_url,
      CURLOPT_HTTPHEADER     => $headers,
      CURLOPT_USERPWD        => $username . ":" . $password,
      CURLOPT_RETURNTRANSFER => true,
      #CURLOPT_CAINFO => "icinga.synyx.coffee.ca.crt", //re-use the icinga2 master ca.crt
      #CURLOPT_SSL_VERIFYHOST => 2,
      #CURLOPT_SSL_VERIFYPEER => 1
      CURLOPT_SSL_VERIFYHOST => 0,
      CURLOPT_SSL_VERIFYPEER => 0,
    ]
  );
  if (!$json = curl_exec($ch)) {
    return ["<pre>Attempt to hit API failed, sorry. Curl said: " . curl_error($ch) . "</pre>"];
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  curl_close($ch);

  if (!$state = json_decode($json, true)) {
    return ["Attempt to hit API failed, sorry (JSON decode failed)"];
  }
  $curl_stats["$hostname:$port"]['objects'] += count($state['results']);

  return $state;
}

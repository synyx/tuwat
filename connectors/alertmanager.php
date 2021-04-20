<?php

/**
 * @param string $url Alertmanager API URL
 * @param string $token_url OAuth2 Token URL
 * @return string | array Error as string
 */
function connectAlertmanager($url, $token_url = null) {
  $hdr = $token_url ? [oauth2_auth_header($url, $token_url)] : [];

  $state = alertmanagerV2Get($url, "alerts", $hdr);

  if (is_string($state)) {
    return $state;
  }

  if (count($state) == 0) {
    return array("url" => array(
      'services' => [],
      'downtimes' => [],
      'current_state' => 1,
    ));
  }

  $host_state = array_reduce($state, function ($hosts, $alert) {
    $labels = $alert['labels'];

    if (isset($labels['severity']) && $labels['severity'] == 'none') {
      return $hosts;
    }

    $hn = implode(':', k8slabels($labels, array('cluster', 'namespace')));
    if (!isset($hosts[$hn])) {
      $hosts[$hn] = array(
        'services' => [],
        'downtimes' => [],
        'current_state' => 0, # "host" is always up
      );
    }

    $state_mapping = array('unprocessed' => 3, 'active' => 1, 'suppressed' => 1);
    $ack_mapping = array('unprocessed' => 0, 'active' => 0, 'suppressed' => 1);

    $startsAt = DateTime::createFromFormat('Y-m-d\TH:i:s+', $alert['startsAt'],  new DateTimeZone('Etc/Zulu'));
    $description = isset($alert['annotations']['description']) ? $alert['annotations']['description'] : '';
    $link = isset($alert['annotations']['runbook']) ? '<a href="'.$alert['annotations']['runbook'].'" target="_blank">&#x1F4D6; Runbook</a>' : '';

    $sn = implode(' ', k8slabels($labels, array('alertname', 'container', 'endpoint', 'pod')));
    $hosts[$hn]['services'][$sn] = array(
      'current_state'                 => $state_mapping[$alert['status']['state']],
      'problem_has_been_acknowledged' => $ack_mapping[$alert['status']['state']],
      'scheduled_downtime_depth'      => 0,
      'notifications_enabled'         => count($alert['status']['silencedBy']) > 0 ? 0 : 1,
      'plugin_output'                 => $description . ' ' . $link,
      'max_attempts'                  => 1,
      'current_attempt'               => 1,
      'state_type'                    => 1,
      'downtimes'                     => [],
      'last_state_change'             => $startsAt->getTimestamp(),
      'labels'                        => $labels,
    );

    return $hosts;
  }, array());

  #print_r($host_state);
  return $host_state;
}

function k8slabels($labels, $keys) {
  $ret = [];
  foreach ($keys as $key) {
    if (isset($labels[$key])) {
      $ret[] = $labels[$key];
    }
  }
  return $ret;
}

function alertmanagerV2Get($url, $endpoint, $headers = []) {
  global $curl_stats;

  $hostname    = parse_url($url, PHP_URL_HOST);
  $scheme      = parse_url($url, PHP_URL_SCHEME);
  $port        = parse_url($url, PHP_URL_PORT);
  $request_url = "$scheme://$hostname:$port/v2/{$endpoint}";
  $headers     = array_merge([
    'Accept: application/json'
  ], $headers);
  $ch          = curl_init();
  curl_setopt_array(
    $ch, [
      CURLOPT_URL            => $request_url,
      CURLOPT_HTTPHEADER     => $headers,
      CURLOPT_RETURNTRANSFER => true,
      CURLOPT_SSL_VERIFYHOST => 0,
    ]
  );
  curl_setopt($ch, CURLOPT_VERBOSE, true);
  if (!$json = curl_exec($ch)) {
    return "<pre>Attempt to hit Alertmanager API failed, sorry. Curl said: " . curl_error($ch) . "</pre>";
  } else {
    $curl_stats["$hostname:$port"] = curl_getinfo($ch);
  }
  curl_close($ch);

  if (!$state = json_decode($json, true)) {
    return "Attempt to parse alertmanager json failed, sorry (JSON decode failed)";
  }
  $curl_stats["$hostname:$port"]['objects'] += count($state);

  return $state;
}

function oauth2_auth_header($url, $token_url) {
  $client_id = parse_url($url, PHP_URL_USER);
  $client_secret = parse_url($url, PHP_URL_PASS);
  $token = oauth2_get_token($token_url, $client_id, $client_secret);
  return "Authorization: Bearer {$token['access_token']}";
}

function oauth2_get_token($token_url, $client_id, $client_secret) {
  $req = array(
    'grant_type' => 'client_credentials',
    'client_id' => $client_id,
    'client_secret' => $client_secret,
  );

  $ch = curl_init();
  curl_setopt($ch, CURLOPT_URL, $token_url);
  curl_setopt($ch, CURLOPT_HEADER, false);
  curl_setopt($ch, CURLOPT_RETURNTRANSFER, true);
  curl_setopt($ch, CURLOPT_POST, true);
  curl_setopt($ch, CURLOPT_POSTFIELDS, http_build_query($req));
  curl_setopt($ch, CURLOPT_VERBOSE, true);
  $body = curl_exec($ch);
  curl_close($ch);

  return json_decode($body, true);
}

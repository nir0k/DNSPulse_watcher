import sys
import requests
import json
import argparse
import time


def create_parser():
    parser = argparse.ArgumentParser(
        description=('Fetch unique label combinations from Metrics time'
                     ' siries data.'))
    parser.add_argument(
        '--host',
        type=str,
        help='Victoria Metrics host, e.g. http://localhost:8428')
    parser.add_argument(
        '--query', type=str, help='PROMQL query to be executed', default='')
    parser.add_argument(
        '--file_path', type=str, help='Faile path for save query result.')
    parser.add_argument(
        '--unique-labels', action='store_true',
        help='Fetch unique label combinations')
    parser.add_argument(
        '--single-host', type=str,
        help='Fetch unique label combinations for single host',
        default='')
    parser.add_argument(
        '--search_criteria',
        type=str,
        help='JSON string of search criteria for matching metrics.')

    return parser


def fetch_time_series_data(host, query):
    url = f'{host}/api/v1/query'
    response = requests.get(url, params={'query': query})
    response.raise_for_status()
    return response.json()


def get_unique_label_combinations(data):
    unique_combinations = set()

    for result in data['data']['result']:
        labels = result['metric']
        labels.pop('rcode', None)

        label_set = frozenset(labels.items())
        unique_combinations.add(label_set)

    return unique_combinations


def fetch_single_value(host, promql_query):
    data = fetch_time_series_data(host, promql_query)
    try:
        metrics = data['data']['result']
        return metrics
    except (IndexError, KeyError):
        raise ValueError(
            'The data structure returned from the query does'
            ' not have the expected format.')


def fetch_metric_value_by_keys_from_file(file_path, search_criteria):
    try:
        with open(file_path, 'r') as file:
            data = json.load(file)
            for metric in data:
                if all(metric.get('metric', {}).get(k) == v
                       for k, v in search_criteria.items()):
                    return metric
            return None
    except FileNotFoundError:
        print(f'The file {file_pah} was not found.')
        return None
    except json.JSONDecodeError:
        print(f'The file {file_path} does not contain valid JSON data.')
        return None


if __name__ == '__main__':
    parser = create_parser()
    args = parser.parse_args()

    output_file = args.file_path

    if args.unique_labels:

        if args.single_host:
            query = (f'{{__name__="dns_resolve",server="{args.single_host}", rcode!=""}}')
        else:
            query = '{__name__="dns_resolve", rcode!=""}'

        try:
            data = fetch_time_series_data(args.host, query)

            unique_combinations = get_unique_label_combinations(data)

            combinations_list = [
                {
                    'server': dict(labels).get('server', ''),
                    'domain': dict(labels).get('domain', ''),
                    'location': dict(labels).get('location', ''),
                    'recursion': dict(labels).get('recursion', ''),
                    'maitanence': dict(labels).get('maitanence', ''),
                    'server_ip': dict(labels).get('server_ip', ''),
                    'server_security_zone': dict(labels).get(
                        'server_security_zone', ''),
                    'site': dict(labels).get('site', ''),
                    'watcher_location': dict(labels).get(
                        'watcher_location', ''),
                    'watcher_security_zone': dict(labels).get(
                        'watcher_security_zone', ''),
                    'zonename': dict(labels).get('zonename', ''),
                    'watcher': dict(labels).get('watcher', '')
                }
                for labels in unique_combinations
            ]

            print(json.dumps(combinations_list, indent=4))

        except requests.exeptions.HTTPError as http_err:
            print(f'HTTP error occured: {http_err}')
        except Exception as err:
            print(f'An error occured: {err}')

    elif args.query:
        try:
            metrics = fetch_single_value(args.host, args.query)
            with open(output_file, 'w') as file:
                json.dump(metrics, file, indent=4)
            print(f'{int(time.time())}')
        except Exception as err:
            print(f'An error occured: {err}')
    elif args.search_criteria:
        try:
            search_criteria = json.loads(args.search_criteria)
        except json.JSONDecodeError:
            print('Invalid JSON for search criteria.')
            sys.exit(1)

        metric = fetch_metric_value_by_keys_from_file(
            output_file, search_criteria)
        if metric:
            print(f'{metric["value"][1]}')
        else:
            print('No metric found matching the provided criteria.')

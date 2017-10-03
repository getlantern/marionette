package marionette

/*




class MarionetteTransition(object):

    def __init__(self, src, dst, action_block, probability, is_error_transition=False):
        self.src_ = src
        self.dst_ = dst
        self.action_block_ = action_block
        self.probability_ = probability
        self.is_error_transition_ = is_error_transition

    def get_src(self):
        return self.src_

    def get_dst(self):
        return self.dst_

    def get_action_block(self):
        return self.action_block_

    def get_probability(self):
        return self.probability_

    def is_error_transition(self):
        return self.is_error_transition_

class MarionetteFormat(object):

    def set_transport(self, transport):
        self.transport_ = transport

    def get_transport(self):
        return self.transport_

    def set_port(self, port):
        self.port_ = port

    def get_port(self):
        return self.port_

    def set_transitions(self, transitions):
        self.transitions_ = transitions

    def get_transitions(self):
        return self.transitions_

    def set_action_blocks(self, action_blocks):
        self.action_blocks_ = action_blocks

    def get_action_blocks(self):
        return self.action_blocks_


def parse(s):
    s = s.strip()

    retval = MarionetteFormat()

    parsed_format = yacc.parse(s)

    retval.set_transport(parsed_format[0])
    retval.set_port(parsed_format[1])
    retval.set_transitions(parsed_format[2])
    retval.set_action_blocks(parsed_format[3])

    return retval

def get_search_dirs():
    dsl_dir = os.path.dirname(os.path.join(__file__))
    dsl_dir = os.path.join(dsl_dir, 'formats')
    dsl_dir = os.path.abspath(dsl_dir)
    cwd_dir = os.path.join(os.getcwd(), 'marionette_tg', 'formats')
    cwd_dir = os.path.abspath(cwd_dir)
    retval = [dsl_dir, # find formats based on location of dsl.py
              cwd_dir, # find formats based on location of marionette_client.exe
              sys.prefix, # find formats based on location of python install
              sys.exec_prefix, # same as above
             ]
    return retval

def get_format_dir():
    retval = None

    search_dirs = get_search_dirs()
    FORMAT_BANNER = '### marionette formats dir ###'
    for cur_dir in search_dirs:
        init_path = os.path.join(cur_dir, '__init__.py')

        # check if __init__ marks our marionette formats dir
        if os.path.exists(init_path):
            with open(init_path) as fh:
                contents = fh.read()
                contents = contents.strip()
                if contents != FORMAT_BANNER:
                    continue
                else:
                    retval = cur_dir
                    break
        else:
            continue

    return retval

def find_mar_files(party, format_name, version=None):
    retval = []

    # get marionette format directory
    format_dir = get_format_dir()

    # check all subdirs unless a version is specified
    if version:
        subdirs = glob.glob(os.path.join(format_dir, version))
    else:
        subdirs = glob.glob(os.path.join(format_dir, '*'))

    # make sure we prefer the most recent format
    subdirs.sort()

    # for each subdir, load our format_name
    formats = {}
    for path in subdirs:
        if os.path.isdir(path):
            conf_path = os.path.join(path, format_name + '.mar')
            if os.path.exists(conf_path):
                if not formats.get(format_name):
                    formats[format_name] = []
                if party == 'client':
                    formats[format_name] = [conf_path]
                elif party == 'server':
                    formats[format_name] += [conf_path]

    for key in formats.keys():
        retval += formats[key]

    return retval

def list_mar_files(party):
    format_dir = get_format_dir()

    subdirs = glob.glob(os.path.join(format_dir,'*'))

    mar_files = []
    for path in subdirs:
        if os.path.isdir(path):
            format_version = os.path.basename(path)

            for root, dirnames, filenames in os.walk(path):
                for filename in fnmatch.filter(filenames, '*.mar'):
                    full_path = os.path.join(root,filename)
                    rel_path = os.path.relpath(full_path, path)
                    format = os.path.splitext(rel_path)[0]
                    mar_file = "%s:%s" % (format,format_version)
                    mar_files.append(mar_file)

    return mar_files

def get_latest_version(party, format_name):
    mar_version = None

    # get marionette format directory
    format_dir = get_format_dir()

    subdirs = glob.glob(os.path.join(format_dir, '*'))

    # make sure we prefer the most recent format
    subdirs.sort()

    # for each subdir, load our format_name
    for path in subdirs:
        if os.path.isdir(path):
            conf_path = os.path.join(path, format_name + '.mar')
            if os.path.exists(conf_path):
                mar_version = path.split('/')[-1]

    return mar_version


def load_all(party, format_name, version=None):
    retval = []

    mar_files = find_mar_files(party, format_name, version)
    if not mar_files:
        raise Exception("Can't find "+format_name)

    for mar_path in mar_files:
        retval.append(load(party, format_name, mar_path))

    return retval


def load(party, format_name, mar_path):

    with open(mar_path) as f:
        mar_str = f.read()

    parsed_format = parse(mar_str)

    first_sender = 'client'
    if format_name in ["ftp_pasv_transfer"]:
        first_sender = "server"

    executable = marionette_tg.executables.pioa.PIOA(party, first_sender)
    executable.set_transport_protocol(parsed_format.get_transport())
    executable.set_port(parsed_format.get_port())
    executable.set_local(
        "model_uuid", get_model_uuid(mar_str))

    for transition in parsed_format.get_transitions():
        executable.add_state(transition.get_src())
        executable.add_state(transition.get_dst())
        executable.states_[
            transition.get_src()].add_transition(
            transition.get_dst(),
            transition.get_action_block(),
            transition.get_probability())
        if transition.is_error_transition():
            executable.states_[
                transition.get_src()].set_error_transition(transition.get_dst())

    actions = []
    for action in parsed_format.get_action_blocks():
        actions.append(action)
        complementary_action = copy.deepcopy(action)
        if action.get_module() in ['fte', 'tg']:
            if action.get_method() == 'send':
                complementary_method = 'recv'
            elif action.get_method() == 'send_async':
                complementary_method = 'recv_async'
            complementary_party = 'client' if action.get_party(
            ) == 'server' else 'server'

            complementary_action.set_method(complementary_method)
            complementary_action.set_party(complementary_party)

            actions.append(complementary_action)
        elif action.get_module() in ['io']:
            complementary_method = 'gets' if action.get_method(
            ) == 'puts' else 'puts'
            complementary_party = 'client' if action.get_party(
            ) == 'server' else 'server'

            complementary_action.set_method(complementary_method)
            complementary_action.set_party(complementary_party)

            actions.append(complementary_action)

    executable.actions_ = actions
    executable.do_precomputations()

    if executable.states_.get("end"):
        executable.add_state("dead")
        executable.states_["end"].add_transition("dead", None, 1)
        executable.states_["dead"].add_transition("dead", None, 1)

    return executable


def get_model_uuid(format_str):
    m = hashlib.md5()
    m.update(format_str)
    bytes = m.digest()
    return fte.bit_ops.bytes_to_long(bytes[:4])
*/
